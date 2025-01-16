package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

var ErrNoRecords = errors.New("memory: no records error")

type Storage interface {
	Get(mType, mName string) (*models.Metric, error)
	Set(ctx context.Context, m *models.Metric) error
}

type Memory struct {
	storage map[string]map[string]string
	mutex   sync.Mutex
	lg      *logging.ZapLogger
}

func NewMemoryStorage(lg *logging.ZapLogger) *Memory {
	storage := make(map[string]map[string]string)

	return &Memory{
		lg:      lg,
		storage: storage,
		mutex:   sync.Mutex{},
	}
}

func (s *Memory) Set(ctx context.Context, m *models.Metric) error {
	ctx = s.lg.WithContextFields(ctx, zap.String("name", "memoty_storage"))
	s.mutex.Lock()
	defer s.mutex.Unlock()

	mTypeStorage, ok := s.storage[m.Type]
	if !ok {
		mTypeStorage = make(map[string]string)
		s.storage[m.Type] = mTypeStorage
	}

	s.lg.DebugCtx(ctx,
		"add metric to storage",
		zap.String("type", m.Type),
		zap.String("name", m.Name),
		zap.String("value", m.Value),
	)
	mTypeStorage[m.Name] = m.Value
	return nil
}

func (s *Memory) Get(mType, mName string) (*models.Metric, error) {
	mTypeStorage, ok := s.storage[mType]
	if !ok {
		return nil, fmt.Errorf("storage/memory: Got type: %v, name: %v - type %w", mType, mName, ErrNoRecords)
	}

	val, ok := mTypeStorage[mName]
	if !ok {
		return nil, fmt.Errorf("storage/memory: Got type: %v, name: %v - value %w", mType, mName, ErrNoRecords)
	}

	return &models.Metric{Name: mName, Type: mType, Value: val}, nil
}
