package pubsub

import (
	"bufio"
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type Message struct {
	MName  string `json:"name"`
	MType  string `json:"type"`
	MValue any    `json:"value"`
}

type Publisher struct {
	ctx context.Context
	lg  *logging.ZapLogger
	ch  chan *Message
}

func newPublisher(lg *logging.ZapLogger, ch chan *Message) *Publisher {
	return &Publisher{
		ctx: lg.WithContextFields(context.Background(), zap.String("name", "publisher")),
		lg:  lg,
		ch:  ch,
	}
}

func (p *Publisher) Push(m *Message) {
	p.lg.DebugCtx(p.ctx, "publish message", zap.Any("message", m))
	p.ch <- m
}

type Queue struct {
	l   *list.List
	mtx *sync.Mutex
}

func newQueue() *Queue {
	return &Queue{
		l:   list.New(),
		mtx: &sync.Mutex{},
	}
}

type mAction int

const (
	push mAction = iota
	pop
)

func (q *Queue) push(m *Message) {
	q.mutate(push, m)
}

func (q *Queue) pop() *list.Element {
	return q.mutate(pop)
}

func (q *Queue) len() int {
	return q.l.Len()
}

func (q *Queue) mutate(action mAction, m ...*Message) *list.Element {
	q.mtx.Lock()
	defer q.mtx.Unlock()

	switch action {
	case push:
		return q.l.PushFront(*m[0])
	case pop:
		if el := q.l.Back(); el != nil {
			q.l.Remove(el)
			return el
		}
	}

	return nil
}

type Subscriber struct {
	fStoragePath  string
	lg            *logging.ZapLogger
	ch            chan *Message
	q             *Queue
	storeInterval time.Duration
	mtx           *sync.Mutex
	dw            *dataWriter
}

type dataWriter struct {
	mtx *sync.Mutex
	to  io.WriteCloser
	w   *bufio.Writer
}

func newDataWriter(to io.WriteCloser) *dataWriter {
	return &dataWriter{
		to:  to,
		mtx: &sync.Mutex{},
		w:   bufio.NewWriter(to),
	}
}

func (dw *dataWriter) Close() error {
	return dw.to.Close()
}

func newSubscriber(lg *logging.ZapLogger, ch chan *Message, storeInterval time.Duration, fStoragePath string, dw *dataWriter) *Subscriber {
	return &Subscriber{
		lg:            lg,
		ch:            ch,
		q:             newQueue(),
		storeInterval: storeInterval,
		fStoragePath:  fStoragePath,
		mtx:           &sync.Mutex{},
		dw:            dw,
	}
}

func (s *Subscriber) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		s.startConsumer(ctx, wg)
		s.startScheduller(ctx, wg)
	}()
}

func (s *Subscriber) startScheduller(ctx context.Context, wg *sync.WaitGroup) {
	ctx = s.lg.WithContextFields(ctx, zap.String("actor", "metrics_writer_scheduller"))

	if s.isSync() {
		s.lg.DebugCtx(ctx, "skip")
		return
	}

	wg.Add(1)
	go func() {
		defer s.dw.Close()
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				ctx := s.lg.WithContextFields(ctx, zap.String("uuid", uuid.NewV4().String()))
				s.lg.DebugCtx(ctx, "graceful shutdown", zap.String("stage", "start"))
				s.unqueue(ctx)
				s.lg.DebugCtx(ctx, "graceful shutdown", zap.String("stage", "finished"))
				return
			case <-time.After(s.storeInterval):
				ctx := s.lg.WithContextFields(ctx, zap.String("uuid", uuid.NewV4().String()))
				s.lg.DebugCtx(ctx, "scheduler", zap.String("stage", "start"))
				s.unqueue(ctx)
				s.lg.DebugCtx(ctx, "scheduler", zap.String("stage", "finised"))
			}
		}
	}()
}

func (s *Subscriber) startConsumer(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		ctx = s.lg.WithContextFields(ctx, zap.String("actor", "metrics_writer_consumer"))

		for {
			select {
			case <-ctx.Done():
				s.lg.DebugCtx(ctx, "graceful shutdown")
				return
			case m := <-s.ch:
				if s.isSync() {
					s.lg.DebugCtx(ctx, "consumed message, do sync append", zap.Any("message", m))
					s.appendMetrics(ctx, m)
				} else {
					s.lg.DebugCtx(ctx, "consumed message, do publish to queue", zap.Any("message", m))
					s.q.push(m)
				}
			}
		}
	}()
}

func (s *Subscriber) unqueue(ctx context.Context) {
	metrics := make([]*Message, 0)

	for {
		el := s.q.pop()
		if el == nil {
			break
		}

		s.lg.DebugCtx(ctx, "Pop element from queue", zap.Any("element", el.Value))
		if m, ok := el.Value.(Message); ok {
			metrics = append(metrics, &m)
		} else {
			s.lg.ErrorCtx(ctx, "invalid queue element", zap.Any("element", m))
			continue
		}
	}

	s.appendMetrics(ctx, metrics...)
	s.lg.DebugCtx(ctx, "queue is empty")
}

func (s *Subscriber) isSync() bool {
	return s.storeInterval == 0
}

type PubSub struct {
	ch chan *Message
	Pb *Publisher
	Sb *Subscriber
}

var ErrUnexpectedWriter = errors.New("got unexpected writer type")

func NewPubSub(lg *logging.ZapLogger, cfg config.Config, to io.WriteCloser) *PubSub {
	ch := make(chan *Message)
	pb := newPublisher(
		lg,
		ch,
	)
	sb := newSubscriber(
		lg,
		ch,
		time.Duration(cfg.StoreInterval)*time.Second,
		cfg.FileStoragePath,
		newDataWriter(to),
	)

	return &PubSub{
		ch: ch,
		Pb: pb,
		Sb: sb,
	}
}

func (s *Subscriber) appendMetrics(ctx context.Context, messages ...*Message) {
	if len(messages) == 0 {
		return
	}
	for _, m := range messages {
		b, err := json.Marshal(m)
		if err != nil {
			s.lg.ErrorCtx(ctx, "marshal metric failed", zap.Error(err), zap.Any("metric", m))
			continue
		}
		if _, err = s.dw.w.Write(b); err != nil {
			s.lg.ErrorCtx(ctx, "write metric failed", zap.Error(err), zap.Any("metric", m))
			continue
		}

		if err := s.dw.w.WriteByte('\n'); err != nil {
			s.lg.ErrorCtx(ctx, "write \n failed", zap.Error(err), zap.Any("metric", m))
			continue
		}

		s.dw.w.Flush()
	}
}
