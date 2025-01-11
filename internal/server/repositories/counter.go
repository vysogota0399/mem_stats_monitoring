package repositories

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

type Counter struct {
	storage storage.Storage
	Records []models.Counter
}

func NewCounter(strg storage.Storage) Counter {
	return Counter{
		storage: strg,
		Records: make([]models.Counter, 0),
	}
}

func (c *Counter) Craete(ctx context.Context, record *models.Counter) (*models.Counter, error) {
	var counter int64
	last, err := c.Last(ctx, record.Name)
	if err != nil {
		if err != storage.ErrNoRecords {
			return nil, err
		}
	} else {
		counter = last.Value
	}

	record.Value += counter
	if s, ok := c.storage.(storage.DBAble); ok {
		return c.pushToDB(ctx, s, record)
	}

	if err := c.storage.Push(models.CounterType, record.Name, record); err != nil {
		return nil, err
	}

	return record, nil
}

func (c *Counter) pushToDB(ctx context.Context, s storage.DBAble, rec *models.Counter) (*models.Counter, error) {
	res, err := s.DB().ExecContext(
		ctx,
		`
			insert into counters(name, value)
			values ($1, %2)
		`,
		rec.Name,
		rec.Value,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	rec.ID = id
	return rec, nil
}

func (c Counter) Last(ctx context.Context, mName string) (*models.Counter, error) {
	if s, ok := c.storage.(storage.DBAble); ok {
		return c.lastFromDB(ctx, s, mName)
	}

	return c.lastFromMem(mName)
}

func (c Counter) lastFromDB(ctx context.Context, s storage.DBAble, mName string) (*models.Counter, error) {
	row := s.DB().QueryRowContext(
		ctx,
		`
		select value, id
		from counters
		where name = $1
		order by created_at desc
		limit 1`, mName)
	cntr := &models.Counter{Name: mName}

	if err := row.Scan(cntr.Value, cntr.ID); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNoRecords
		} else {
			return nil, err
		}
	}

	return cntr, nil
}

func (c Counter) lastFromMem(mName string) (*models.Counter, error) {
	record, err := c.storage.Last(models.CounterType, mName)
	if err != nil {
		return nil, err
	}

	var Counter models.Counter

	if err := json.Unmarshal([]byte(record), &Counter); err != nil {
		return nil, err
	}

	return &Counter, nil
}

func (c Counter) All() map[string][]models.Counter { //nolint:dupl // :/
	records := map[string][]models.Counter{}
	mNames, ok := c.storage.All()[models.CounterType]
	if !ok {
		return records
	}

	for name, values := range mNames {
		count := len(values)
		collection := make([]models.Counter, count)
		for i := 0; i < count; i++ {
			collection[i] = models.Counter{}
			if err := json.Unmarshal([]byte(values[i]), &collection[i]); err != nil {
				continue
			}
		}
		records[name] = collection
	}

	return records
}
