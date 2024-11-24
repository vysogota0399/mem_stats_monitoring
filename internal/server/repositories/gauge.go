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
		mType:   "guage",
		Records: make([]models.Gauge, 0),
	}
}

func (g *Gauge) Craete(record models.Gauge) error {
	return g.storage.Push(g.mType, record.Name, record)
}

func (g Gauge) Last(mName string) (models.Gauge, error) {
	record, err := g.storage.Last(g.mType, mName)
	if err != nil {
		return models.Gauge{}, err
	}

	var gauge models.Gauge

	if err := json.Unmarshal([]byte(record), &gauge); err != nil {
		return models.Gauge{}, err
	}

	return gauge, nil
}
