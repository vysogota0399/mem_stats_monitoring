package storage

import (
	"errors"
	"sync"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

var ErrNoRecords = errors.New("memory: no records error")

type Storage interface {
	Get(mType, mName string) (*models.Metric, error)
	Set(m *models.Metric) error
}

type Memory struct {
	storage map[string]map[string]string
	mutex   sync.Mutex
	logger  utils.Logger
}

func NewMemoryStorage() *Memory {
	logger := utils.InitLogger("[storage]")
	storage := make(map[string]map[string]string)

	return &Memory{
		logger:  logger,
		storage: storage,
		mutex:   sync.Mutex{},
	}
}

func (s *Memory) Set(m *models.Metric) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	mTypeStorage, ok := s.storage[m.Type]
	if !ok {
		s.logger.Printf("Add metric type(%s) to storage", m.Type)
		mTypeStorage = make(map[string]string)
		s.storage[m.Type] = mTypeStorage
	}
	mTypeStorage[m.Name] = m.Value
	return nil
}

func (s *Memory) Get(mType, mName string) (*models.Metric, error) {
	mTypeStorage, ok := s.storage[mType]
	if !ok {
		s.logger.Printf("Got type: %v, name: %v, result: type not found", mType, mName)
		return nil, ErrNoRecords
	}

	val, ok := mTypeStorage[mName]
	if !ok {
		s.logger.Printf("Got type: %v, name: %v\nResult: value not found", mType, mName)
		return nil, ErrNoRecords
	}

	return &models.Metric{Name: mName, Type: mType, Value: val}, nil
}
