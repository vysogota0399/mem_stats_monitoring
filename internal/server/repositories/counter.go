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

func (c *Counter) Craete(record models.Counter) error {
	var counter int64
	last, err := c.Last(record.Name)
	if err != nil && err != storage.ErrNoRecords {
		return err
	} else {
		counter = last.Value
	}

	record.Value += counter
	return c.storage.Push(c.mType, record.Name, record)
}

func (c Counter) Last(mName string) (models.Counter, error) {
	record, err := c.storage.Last(c.mType, mName)
	if err != nil {
		return models.Counter{}, err
	}

	var Counter models.Counter

	if err := json.Unmarshal([]byte(record), &Counter); err != nil {
		return models.Counter{}, err
	}

	return Counter, nil
}
