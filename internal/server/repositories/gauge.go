package repositories

import (
	"encoding/json"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

type Gauge struct {
	storage storage.Storage
	Records []models.Gauge
	mType   string
}

func NewGauge(storage storage.Storage) Gauge {
	return Gauge{
		storage: storage,
		mType:   "gauge",
		Records: make([]models.Gauge, 0),
	}
}

func (g *Gauge) Craete(record models.Gauge) (models.Gauge, error) {
	if err := g.storage.Push(g.mType, record.Name, record); err != nil {
		return record, err
	}

	return record, nil
}

func (g Gauge) Last(mName string) (*models.Gauge, error) {
	record, err := g.storage.Last(g.mType, mName)
	if err != nil {
		return nil, err
	}

	var gauge models.Gauge

	if err := json.Unmarshal([]byte(record), &gauge); err != nil {
		return nil, err
	}

	return &gauge, nil
}

func (g Gauge) All() map[string][]models.Gauge { //nolint:dupl // :/
	records := map[string][]models.Gauge{}
	mNames, ok := g.storage.All()[g.mType]
	if !ok {
		return records
	}

	for name, values := range mNames {
		count := len(values)
		collection := make([]models.Gauge, count)
		for i := 0; i < count; i++ {
			collection[i] = models.Gauge{}
			if err := json.Unmarshal([]byte(values[i]), &collection[i]); err != nil {
				continue
			}
		}
		records[name] = collection
	}

	return records
}
