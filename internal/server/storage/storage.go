package storage

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

type Storage interface {
	Last(mType, mName string) (string, error)
	Push(mType, mName string, val any) error
	All() map[string]map[string][]string
}

var ErrNoRecords = errors.New("memory: no records error")

// storage
//
//	{
//		"gauge": {
//			"fiz": [1,1,1,1]
//		},
//		"counter": {
//			"baz": [1,2,3,4]
//		}
//	}
type Memory struct {
	storage map[string]map[string][]string
	logger  utils.Logger
	mutex   sync.Mutex
}

func NewMemStorageWithData(storage map[string]map[string][]string, logger utils.Logger) *Memory {
	return &Memory{storage: storage, logger: logger}
}

func New() Storage {
	return NewMemory()
}

func NewMemory() *Memory {
	return &Memory{
		logger:  utils.InitLogger("[storage]"),
		storage: make(map[string]map[string][]string),
		mutex:   sync.Mutex{},
	}
}

func (m *Memory) Push(mType, mName string, val any) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.logger.Printf("Push type %v, name %v value: %v", mType, mName, val)
	mTypeStorage, ok := m.storage[mType]
	if !ok {
		m.logger.Printf("Add metric type(%s) to storage", mType)
		mTypeStorage = make(map[string][]string)
		m.storage[mType] = mTypeStorage
	}

	valuesStorage, ok := mTypeStorage[mName]
	if !ok {
		m.logger.Printf("Add metric name(%s) to storage", mName)
		valuesStorage = make([]string, 1)
		mTypeStorage[mName] = valuesStorage
	}

	jsonVal, err := json.Marshal(val)
	if err != nil {
		m.logger.Printf("JSON marshal error: %v", err)
		return err
	}
	strVal := string(jsonVal)
	m.logger.Printf("Add %s to storage %s->%s", strVal, mType, mName)

	mTypeStorage[mName] = append(valuesStorage, strVal)
	m.logger.Printf("Storage updated\n%v", m.storage)
	return nil
}

func (m *Memory) Last(mType, mName string) (string, error) {
	mTypeStorage, ok := m.All()[mType]
	if !ok {
		m.logger.Printf("Got type: %v, name: %v\nResult: type not found", mType, mName)
		return "", ErrNoRecords
	}

	valuesStorage, ok := mTypeStorage[mName]
	if !ok {
		m.logger.Printf("Got type: %v, name: %v\nResult: name not found", mType, mName)
		return "", ErrNoRecords
	}

	if len(valuesStorage) == 0 {
		m.logger.Printf("Got type: %v, name: %v\nResult: storage is empty", mType, mName)
		return "", ErrNoRecords
	}

	result := valuesStorage[len(valuesStorage)-1]
	m.logger.Printf("Got type: %v, name: %v\nResult: %v", mType, mName, result)
	return result, nil
}

func (m *Memory) All() map[string]map[string][]string {
	dist := map[string]map[string][]string{}
	for mType, mNames := range m.storage {
		dist[mType] = map[string][]string{}
		for mName, values := range mNames {
			lenValues := len(values)
			newValues := make([]string, lenValues)
			dist[mType][mName] = newValues
			for i := 0; i < lenValues; i++ {
				newValues[i] = values[i]
			}
		}
	}

	return dist
}
