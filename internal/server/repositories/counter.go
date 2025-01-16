package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

type Counter struct {
	storage storage.Storage
	Records []models.Counter
}

func NewCounter(strg storage.Storage) *Counter {
	return &Counter{
		storage: strg,
		Records: make([]models.Counter, 0),
	}
}

func (c *Counter) Craete(ctx context.Context, record *models.Counter) (*models.Counter, error) {
	record, err := c.processRec(ctx, record)
	if err != nil {
		return nil, err
	}

	if s, ok := c.storage.(storage.DBAble); ok {
		return c.pushToDB(ctx, s, record)
	}

	if err := c.storage.Push(models.CounterType, record.Name, record); err != nil {
		return nil, err
	}

	return record, nil
}

func (c *Counter) processRec(ctx context.Context, record *models.Counter) (*models.Counter, error) {
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
	return record, nil
}

func (c *Counter) pushToDB(ctx context.Context, s storage.DBAble, rec *models.Counter) (*models.Counter, error) {
	if err := s.DB().QueryRowContext(
		ctx,
		`
			insert into counters(name, value)
			values ($1, $2)
			returning id
		`,
		rec.Name,
		rec.Value,
	).Scan(&rec.ID); err != nil {
		return nil, err
	}

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
		order by id desc
		limit 1`, mName)
	cntr := &models.Counter{Name: mName}

	if err := row.Scan(&cntr.Value, &cntr.ID); err != nil {
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

func (c *Counter) SaveCollection(ctx context.Context, coll []models.Counter) ([]models.Counter, error) {
	if s, ok := c.storage.(storage.DBAble); ok {
		return c.saveCollToDB(ctx, s, coll)
	}

	return c.saveCollToMem(coll)
}

func (c *Counter) saveCollToDB(ctx context.Context, s storage.DBAble, coll []models.Counter) ([]models.Counter, error) {
	var names []string
	for _, n := range coll {
		names = append(names, n.Name)
	}

	records, err := c.SearchByName(ctx, names)
	if err != nil {
		return nil, err
	}

	tx, err := s.DB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Commit()

	for _, rec := range coll {
		if found, ok := records[rec.Name]; ok {
			rec.Value += found.Value
		} else {
			records[rec.Name] = rec
		}

		res := tx.QueryRowContext(
			ctx,
			`
			insert into counters(name, value)
			values ($1, $2)
			returning id
			`,
			rec.Name,
			rec.Value,
		)
		if err := res.Scan(&rec.ID); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("internal/server/repositories/counter.go: record scan failed error %w", err)
		}
	}

	return coll, nil
}

func (c *Counter) saveCollToMem(coll []models.Counter) ([]models.Counter, error) {
	for _, rec := range coll {
		if err := c.storage.Push(models.CounterType, rec.Name, &rec); err != nil {
			return nil, err
		}
	}
	return coll, nil
}

func (c *Counter) SearchByName(ctx context.Context, names []string) (map[string]models.Counter, error) {
	if s, ok := c.storage.(storage.DBAble); ok {
		return c.searchSumByNameInDB(ctx, s, names)
	}

	return c.searchByNameInMem(ctx, names)
}

func (c *Counter) searchSumByNameInDB(ctx context.Context, s storage.DBAble, names []string) (map[string]models.Counter, error) {
	rows, err := s.DB().QueryContext(ctx,
		`
			select name, sum(value)
			from counters
			where name = any($1)
			group by name;
	`,
		names,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make(map[string]models.Counter)
	for rows.Next() {
		rec := models.Counter{}
		if err := rows.Scan(&rec.Name, &rec.Value); err != nil {
			return nil, err
		}

		records[rec.Name] = rec
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (c *Counter) searchByNameInMem(ctx context.Context, names []string) (map[string]models.Counter, error) {
	records := make(map[string]models.Counter)
	for _, name := range names {
		rec, err := c.Last(ctx, name)
		if err != nil {
			return nil, err
		}

		records[rec.Name] = *rec
	}

	return records, nil
}
