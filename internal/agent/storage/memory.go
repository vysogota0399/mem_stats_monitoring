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
	Set(m *models.Metric) error
}

type Memory struct {
	storage map[string]map[string]string
	mutex   sync.Mutex
	lg      *logging.ZapLogger
	ctx     context.Context
}

func NewMemoryStorage(ctx context.Context, lg *logging.ZapLogger) *Memory {
	storage := make(map[string]map[string]string)

	return &Memory{
		ctx:     ctx,
		lg:      lg,
		storage: storage,
		mutex:   sync.Mutex{},
	}
}

func (s *Memory) Set(m *models.Metric) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	mTypeStorage, ok := s.storage[m.Type]
	if !ok {
		mTypeStorage = make(map[string]string)
		s.storage[m.Type] = mTypeStorage
	}

	s.lg.DebugCtx(s.ctx,
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
