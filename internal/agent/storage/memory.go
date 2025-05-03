package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

var ErrNoRecords = errors.New("memory: no records error")

type Storage interface {
	Get(to *models.Metric) error
	Set(ctx context.Context, m *models.Metric) error
}

type Memory struct {
	storage map[string]map[string]string
	mutex   sync.RWMutex
	lg      *logging.ZapLogger
}

func NewMemoryStorage(lg *logging.ZapLogger) *Memory {
	storage := make(map[string]map[string]string)

	return &Memory{
		lg:      lg,
		storage: storage,
		mutex:   sync.RWMutex{},
	}
}

func (s *Memory) Set(ctx context.Context, m *models.Metric) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	mTypeStorage, ok := s.storage[m.Type]
	if !ok {
		mTypeStorage = make(map[string]string)
		s.storage[m.Type] = mTypeStorage
	}

	mTypeStorage[m.Name] = m.Value
	return nil
}

func (s *Memory) Get(to *models.Metric) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	mType := to.Type
	mName := to.Name

	mTypeStorage, ok := s.storage[mType]
	if !ok {
		return fmt.Errorf("storage/memory: Got type: %v, name: %v - type %w", mType, mName, ErrNoRecords)
	}

	val, ok := mTypeStorage[mName]
	if !ok {
		return fmt.Errorf("storage/memory: Got type: %v, name: %v - value %w", mType, mName, ErrNoRecords)
	}

	to.Value = val
	return nil
}
