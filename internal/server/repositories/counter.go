package repositories

import (
	"encoding/json"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

type Counter struct {
	storage storage.Storage
	Records []models.Counter
	mType   string
}

func NewCounter(storage storage.Storage) Counter {
	return Counter{
		storage: storage,
		mType:   "counter",
		Records: make([]models.Counter, 0),
	}
}

func (c *Counter) Craete(record models.Counter) (models.Counter, error) {
	var counter int64
	last, err := c.Last(record.Name)
	if err != nil {
		if err != storage.ErrNoRecords {
			return record, err
		}
	} else {
		counter = last.Value
	}

	record.Value += counter
	if err := c.storage.Push(c.mType, record.Name, record); err != nil {
		return record, err
	}

	return record, nil
}

func (c Counter) Last(mName string) (*models.Counter, error) {
	record, err := c.storage.Last(c.mType, mName)
	if err != nil {
		return nil, err
	}

	var Counter models.Counter

	if err := json.Unmarshal([]byte(record), &Counter); err != nil {
		return nil, err
	}

	return &Counter, nil
}
