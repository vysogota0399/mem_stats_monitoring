package dump

import (
	"bufio"
	"bytes"
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type DumpMessage struct {
	MType  string `json:"type"`
	MName  string `json:"name"`
	MValue any    `json:"value"`
}

type MetricsDumper struct {
	lg              *logging.ZapLogger
	pb              publisher
	sb              subscriber
	fileStoragePath string
	doneCh          chan struct{}
}

func NewMetricsDumper(lg *logging.ZapLogger, cfg *config.Config) *MetricsDumper {
	ch := make(chan DumpMessage)
	doneCh := make(chan struct{})
	return &MetricsDumper{
		fileStoragePath: cfg.FileStoragePath,
		lg:              lg,
		doneCh:          doneCh,
		pb:              newPublisher(lg, ch),
		sb: newSubscriber(
			lg,
			ch,
			doneCh,
			time.Duration(cfg.StoreInterval)*time.Second,
		),
	}
}

const rwfmode = 0644

func (d *MetricsDumper) Start(ctx context.Context) error {
	target, err := os.OpenFile(d.fileStoragePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, rwfmode)
	if err != nil {
		return fmt.Errorf("metrics_dumper: failed to open file %w", err)
	}

	d.start(ctx, target)

	return nil
}

func (d *MetricsDumper) start(ctx context.Context, target io.WriteCloser) {
	dw := newDataWriter(target)

	if d.sb.isSync() {
		go d.sb.startSyncConsumer(
			d.lg.WithContextFields(ctx, zap.String("actor", "metrics_dumper_consumer")),
			dw,
		)
		return
	}

	go d.sb.startAsyncConsumer(
		d.lg.WithContextFields(ctx, zap.String("actor", "metrics_dumper_consumer")),
	)

	go d.sb.startScheduller(
		d.lg.WithContextFields(ctx, zap.String("actor", "metrics_dumper_scheduller")),
		dw,
	)
}

func (d *MetricsDumper) Stop(ctx context.Context) {
	d.lg.DebugCtx(ctx, "stop metrics dumper")
	defer d.lg.DebugCtx(ctx, "metrics dumper stopped")

	d.pb.close()
	<-d.doneCh
}

func (d *MetricsDumper) Dump(ctx context.Context, m DumpMessage) {
	d.pb.push(ctx, m)
}

type dataWriter struct {
	mtx sync.Mutex
	to  io.WriteCloser
	w   *bufio.Writer
}

func newDataWriter(to io.WriteCloser) *dataWriter {
	return &dataWriter{
		to:  to,
		mtx: sync.Mutex{},
		w:   bufio.NewWriter(to),
	}
}

func (dw *dataWriter) Close() error {
	return dw.to.Close()
}

func (dw *dataWriter) Write(b []byte) (int, error) {
	dw.mtx.Lock()
	defer dw.mtx.Unlock()
	return dw.w.Write(b)
}

func (dw *dataWriter) Flush() error {
	dw.mtx.Lock()
	defer dw.mtx.Unlock()

	return dw.w.Flush()
}

type publisher struct {
	lg *logging.ZapLogger
	ch chan DumpMessage
}

func newPublisher(lg *logging.ZapLogger, ch chan DumpMessage) publisher {
	return publisher{lg: lg, ch: ch}
}

func (p *publisher) push(ctx context.Context, m DumpMessage) {
	p.lg.DebugCtx(ctx, "publish message", zap.Any("message", m))
	p.ch <- m
}

func (p *publisher) close() {
	close(p.ch)
}

type dumpPool struct {
	pool sync.Pool
}

func newDumpPool() *dumpPool {
	return &dumpPool{
		pool: sync.Pool{
			New: func() any {
				return &bytes.Buffer{}
			},
		},
	}
}

func (p *dumpPool) get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

func (p *dumpPool) put(b *bytes.Buffer) {
	for i := range b.Bytes() {
		b.Bytes()[i] = 0
	}
	b.Reset()
	p.pool.Put(b)
}

type subscriber struct {
	lg            *logging.ZapLogger
	ch            chan DumpMessage
	storeInterval time.Duration
	q             *Queue
	doneCh        chan struct{}

	pool *dumpPool
}

func newSubscriber(lg *logging.ZapLogger, ch chan DumpMessage, doneCh chan struct{}, storeInterval time.Duration) subscriber {
	return subscriber{lg: lg, ch: ch, q: newQueue(), storeInterval: storeInterval, doneCh: doneCh, pool: newDumpPool()}
}

func (s *subscriber) isSync() bool {
	res := s.storeInterval == time.Duration(0)
	s.lg.DebugCtx(context.Background(), "check if sync", zap.Duration("storeInterval", s.storeInterval), zap.Bool("result", res))
	return res
}

func (s *subscriber) startSyncConsumer(ctx context.Context, dw *dataWriter) {
	syncCtx := s.lg.WithContextFields(ctx, zap.String("actor", "sync_consumer"))
	defer func() {
		if err := dw.Close(); err != nil {
			s.lg.ErrorCtx(ctx, "metrics_dumper: close dw error", zap.Error(err))
		}
	}()

	defer func() {
		s.lg.DebugCtx(syncCtx, "finish")
		close(s.doneCh)
	}()

	for m := range s.ch {
		s.lg.DebugCtx(syncCtx, "consumed message, do sync append", zap.Any("message", m))
		if err := s.dump(syncCtx, dw, m); err != nil {
			s.lg.ErrorCtx(syncCtx, "failed to append metrics", zap.Error(err))
		}
	}

	s.lg.DebugCtx(syncCtx, "channel closed")
}

func (s *subscriber) startAsyncConsumer(ctx context.Context) {
	asyncCtx := s.lg.WithContextFields(ctx, zap.String("actor", "async_consumer"))
	defer s.q.close()

	for m := range s.ch {
		s.lg.DebugCtx(asyncCtx, "consumed message, do publish to queue", zap.Any("message", m))
		s.q.push(m)
	}

	s.lg.DebugCtx(asyncCtx, "channel closed, close the queue")
}

func (s *subscriber) dump(ctx context.Context, dw *dataWriter, messages ...DumpMessage) error {
	dumpCtx := s.lg.WithContextFields(ctx, zap.String("method", "dump"))
	s.lg.DebugCtx(dumpCtx, "start")
	defer s.lg.DebugCtx(dumpCtx, "finished")

	if len(messages) == 0 {
		return nil
	}

	for _, m := range messages {
		s.lg.DebugCtx(dumpCtx, "dumping message", zap.Any("message", m))

		buff := s.pool.get()

		if err := json.NewEncoder(buff).Encode(m); err != nil {
			return fmt.Errorf("metrics_dumper: failed to encode message: %w", err)
		}
		if _, err := dw.Write(buff.Bytes()); err != nil {
			return fmt.Errorf("metrics_dumper: failed to write message: %w", err)
		}

		s.pool.put(buff)
	}

	if err := dw.Flush(); err != nil {
		return fmt.Errorf("metrics_dumper: failed to flush: %w", err)
	}

	return nil
}

func (s *subscriber) startScheduller(ctx context.Context, dw *dataWriter) {
	schedulerCtx := s.lg.WithContextFields(ctx, zap.String("actor", "metrics_dumper_scheduller"))
	defer func() {
		if err := dw.Close(); err != nil {
			s.lg.ErrorCtx(ctx, "metrics_dumper: dw close error", zap.Error(err))
		}
	}()
	for {
		select {
		case <-time.After(s.storeInterval):
			s.lg.DebugCtx(schedulerCtx, "scheduler", zap.String("stage", "start"))
			s.unqueue(schedulerCtx, dw)
			s.lg.DebugCtx(schedulerCtx, "scheduler", zap.String("stage", "finised"))
		case <-s.q.doneCh:
			s.lg.DebugCtx(schedulerCtx, "queue is closed")
			s.unqueue(schedulerCtx, dw)
			s.lg.DebugCtx(schedulerCtx, "scheduler finished")
			close(s.doneCh)
			return
		}
	}
}

func (s *subscriber) unqueue(ctx context.Context, dw *dataWriter) {
	metrics := make([]DumpMessage, 0, s.q.len())

	for {
		el := s.q.pop()
		if el == nil {
			break
		}

		s.lg.DebugCtx(ctx, "pop element from queue", zap.Any("element", el.Value))
		if m, ok := el.Value.(DumpMessage); ok {
			metrics = append(metrics, m)
		} else {
			s.lg.ErrorCtx(ctx, "invalid queue element", zap.Any("element", m))
			continue
		}
	}

	if err := s.dump(ctx, dw, metrics...); err != nil {
		s.lg.ErrorCtx(ctx, "failed to dump metrics from queue", zap.Error(err))
	}
	s.lg.DebugCtx(ctx, "queue is empty")
}

type Queue struct {
	mtx    sync.RWMutex
	l      *list.List
	doneCh chan struct{}
}

func newQueue() *Queue {
	return &Queue{
		l:      list.New(),
		mtx:    sync.RWMutex{},
		doneCh: make(chan struct{}),
	}
}

func (q *Queue) pop() *list.Element {
	q.mtx.RLock()
	defer q.mtx.RUnlock()

	if el := q.l.Back(); el != nil {
		q.l.Remove(el)
		return el
	}

	return nil
}

func (q *Queue) close() {
	close(q.doneCh)
}

func (q *Queue) push(m DumpMessage) {
	q.mtx.Lock()
	defer q.mtx.Unlock()

	q.l.PushFront(m)
}

func (q *Queue) len() int {
	q.mtx.RLock()
	defer q.mtx.RUnlock()

	return q.l.Len()
}
