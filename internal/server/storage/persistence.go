package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage/pubsub"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const rwfmode fs.FileMode = 0666

var ErrUnexpectedReader = errors.New("got unexpected reader type")

type messageSaver func(*pubsub.Message)
type dataSource func(*PersistentMemory) (io.ReadCloser, error)

type PersistentMemory struct {
	Memory
	ctx          context.Context
	fStoragePath string
	lg           *logging.ZapLogger
	saveMessage  messageSaver
	dSource      dataSource
}

func NewFilePersistentMemory(ctx context.Context, c config.Config, errg *errgroup.Group, lg *logging.ZapLogger) (*PersistentMemory, error) {
	to, err := os.OpenFile(c.FileStoragePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, rwfmode)
	if err != nil {
		return nil, err
	}

	ps := pubsub.NewPubSub(lg, c, to)
	m := &PersistentMemory{
		Memory:       *NewMemory(),
		ctx:          lg.WithContextFields(ctx, zap.String("name", "persistent_storage")),
		lg:           lg,
		saveMessage:  ps.Pb.Push,
		fStoragePath: c.FileStoragePath,
		dSource:      fileSource,
	}

	if c.Restore {
		errg.Go(func() error {
			if _, err := m.restore(); err != nil {
				return fmt.Errorf("persistance: restore failed error %w", err)
			}

			return nil
		})
	}

	errg.Go(func() error {
		ps.Sb.Start(ctx, errg)

		<-m.ctx.Done()
		lg.DebugCtx(ctx, "graceful shutdown")
		return nil
	})

	return m, nil
}

func (m *PersistentMemory) Push(mType, mName string, val any) error {
	if err := m.Memory.Push(mType, mName, val); err != nil {
		return err
	}

	message := pubsub.Message{}
	switch val := val.(type) {
	case *models.Counter:
		counter := val
		message.MName = counter.Name
		message.MType = models.CounterType
		message.MValue = counter.Value
	case *models.Gauge:
		gauge := val
		message.MName = gauge.Name
		message.MType = models.GaugeType
		message.MValue = gauge.Value
	}

	log.Printf("saveCollToMem %+v %T", message, message)
	m.saveMessage(&message)

	return nil
}

type persistentMetric struct {
	MType  string `json:"type"`
	MName  string `json:"name"`
	MValue any    `json:"value"`
}

func (m *PersistentMemory) restore() ([]persistentMetric, error) {
	ctx := m.lg.WithContextFields(m.ctx, zap.String("action", "restore"))
	from, err := m.dSource(m)
	if err != nil {
		return nil, err
	}
	defer from.Close()

	scanner := *bufio.NewScanner(from)
	var sucCntr int64
	var failCntr int64

	metrics := make([]persistentMetric, 0)
	for scanner.Scan() {
		sucCntr++
		var record persistentMetric
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			failCntr++
			m.lg.ErrorCtx(ctx, "unmarshal record failed", zap.String("token", scanner.Text()), zap.Error(err))
			continue
		}

		m.lg.DebugCtx(ctx, "push to store", zap.Any("record", record))
		if err := m.Memory.Push(record.MType, record.MName, record); err != nil {
			failCntr++
			m.lg.ErrorCtx(ctx, "push record failed", zap.Error(err))
			continue
		}

		metrics = append(metrics, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	m.lg.DebugCtx(ctx, "storage restored", zap.Int64("succeeded", sucCntr), zap.Int64("failed", failCntr))
	return metrics, nil
}

func fileSource(m *PersistentMemory) (io.ReadCloser, error) {
	return os.OpenFile(m.fStoragePath, os.O_RDONLY|os.O_CREATE, rwfmode)
}
