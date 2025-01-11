package repositories

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

type Gauge struct {
	storage storage.Storage
	Records []models.Gauge
}

func NewGauge(strg storage.Storage) Gauge {
	return Gauge{
		storage: strg,
		Records: make([]models.Gauge, 0),
	}
}

func (g *Gauge) Craete(ctx context.Context, record *models.Gauge) (*models.Gauge, error) {
	if s, ok := g.storage.(storage.DBAble); ok {
		return g.pushToDB(ctx, s, record)
	}

	if err := g.storage.Push(models.GaugeType, record.Name, record); err != nil {
		return record, err
	}

	return record, nil
}

func (g *Gauge) pushToDB(ctx context.Context, s storage.DBAble, r *models.Gauge) (*models.Gauge, error) {
	res, err := s.DB().ExecContext(
		ctx,
		`
			insert into gauges(name, value)
			values ($1, %2)
		`,
		r.Name,
		r.Value,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	r.ID = id
	return r, nil
}

func (g Gauge) Last(ctx context.Context, mName string) (*models.Gauge, error) {
	if s, ok := g.storage.(storage.DBAble); ok {
		return g.lastFromDB(ctx, s, mName)
	}

	return g.lastFromMem(mName)
}

func (g Gauge) lastFromDB(ctx context.Context, s storage.DBAble, mName string) (*models.Gauge, error) {
	row := s.DB().QueryRowContext(
		ctx,
		`
		select value, id
		from gauges
		where name = $1
		order by created_at desc
		limit 1`, mName)

	gg := &models.Gauge{Name: mName}
	if err := row.Scan(gg.Value, gg.ID); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNoRecords
		} else {
			return nil, err
		}
	}

	return gg, nil
}

func (g Gauge) lastFromMem(mName string) (*models.Gauge, error) {
	record, err := g.storage.Last(models.GaugeType, mName)
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
	mNames, ok := g.storage.All()[models.GaugeType]
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
