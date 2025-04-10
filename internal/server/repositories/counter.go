package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

// Counter отвечает за связь уровня бизнес логики и persistence layer в контексте работы с Counter.
type Counter struct {
	storage storage.Storage
	Records []models.Counter
	lg      *logging.ZapLogger
}

type DBAble interface {
	storage.Storage
	QueryRowContext(ctx context.Context, query string, args any, result storage.ResultFunc) error
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	CommitTx(ctx context.Context, tx *sql.Tx) error
	RollbackTx(ctx context.Context, tx *sql.Tx) error
}

func NewCounter(strg storage.Storage, lg *logging.ZapLogger) *Counter {
	return &Counter{
		storage: strg,
		Records: make([]models.Counter, 0),
		lg:      lg,
	}
}

// Create сохраняет новую запись в хранилище.
func (c *Counter) Create(ctx context.Context, record *models.Counter) (*models.Counter, error) {
	record, err := c.processRec(ctx, record)
	if err != nil {
		return nil, err
	}

	if s, ok := c.storage.(DBAble); ok {
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

func (c *Counter) pushToDB(ctx context.Context, s DBAble, rec *models.Counter) (*models.Counter, error) {
	if err := s.QueryRowContext(
		ctx,
		`
			INSERT INTO counters(name, value)
			VALUES ($1, $2)
			RETURNING id
		`,
		[]any{rec.Name, rec.Value},
		func(rows *sql.Rows) error {
			return rows.Scan(&rec.ID)
		},
	); err != nil {
		return nil, err
	}

	return rec, nil
}

// Last возвращает последнюю запись из хранилища.
func (c Counter) Last(ctx context.Context, mName string) (*models.Counter, error) {
	if s, ok := c.storage.(DBAble); ok {
		return c.lastFromDB(ctx, s, mName)
	}

	return c.lastFromMem(mName)
}

func (c Counter) lastFromDB(ctx context.Context, s DBAble, mName string) (*models.Counter, error) {
	cntr := &models.Counter{Name: mName}

	if err := s.QueryRowContext(
		ctx,
		`
		SELECT value, id
		FROM counters
		WHERE name = $1
		ORDER BY id DESC
		LIMIT 1`,
		[]any{mName},
		func(rows *sql.Rows) error {
			if err := rows.Scan(&cntr.Value, &cntr.ID); err != nil {
				return err
			}
			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("internal/server/repositories/counter.go: query row context error %w", err)
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

// All возвращает все записи из хранилища.
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

// SaveCollection сохраняет пачку записей в хранилище
func (c *Counter) SaveCollection(ctx context.Context, coll []models.Counter) ([]models.Counter, error) {
	if s, ok := c.storage.(DBAble); ok {
		return c.saveCollToDB(ctx, s, coll)
	}

	return c.saveCollToMem(coll)
}

func (c *Counter) saveCollToDB(ctx context.Context, s DBAble, coll []models.Counter) ([]models.Counter, error) {
	names := make([]string, 0, len(coll))

	for _, n := range coll {
		names = append(names, n.Name)
	}

	records, err := c.SearchByName(ctx, names)
	if err != nil {
		return nil, err
	}

	tx, err := s.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if commitErr := s.CommitTx(ctx, tx); commitErr != nil {
			if rollbackErr := s.RollbackTx(ctx, tx); rollbackErr != nil {
				c.lg.ErrorCtx(ctx, "failed to rollback transaction after commit error", zap.Error(rollbackErr))
			}
		}
	}()

	for _, rec := range coll {
		if found, ok := records[rec.Name]; ok {
			rec.Value += found.Value
		} else {
			records[rec.Name] = rec
		}

		if err := s.QueryRowContext(ctx,
			`
			INSERT INTO counters(name, value)
			VALUES ($1, $2)
			RETURNING id
			`,
			[]any{rec.Name, rec.Value},
			func(rows *sql.Rows) error {
				return rows.Scan(&rec.ID)
			},
		); err != nil {
			if rollbackErr := s.RollbackTx(ctx, tx); rollbackErr != nil {
				c.lg.ErrorCtx(ctx, "failed to rollback transaction after query row context error", zap.Error(rollbackErr))
			}

			return nil, fmt.Errorf("internal/server/repositories/counter.go: query row context error %w", err)
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

// SearchByName осуществляет поиск записей по имени.
func (c *Counter) SearchByName(ctx context.Context, names []string) (map[string]models.Counter, error) {
	if s, ok := c.storage.(DBAble); ok {
		return c.searchSumByNameInDB(ctx, s, names)
	}

	return c.searchByNameInMem(ctx, names)
}

func (c *Counter) searchSumByNameInDB(ctx context.Context, s DBAble, names []string) (map[string]models.Counter, error) {
	records := make(map[string]models.Counter, 100)
	err := s.QueryRowContext(ctx,
		`
			SELECT name, sum(value)
			FROM counters
			WHERE name = ANY($1)
			GROUP BY name;
	`,
		[]any{names},
		func(rows *sql.Rows) error {
			for rows.Next() {
				rec := models.Counter{}
				if err := rows.Scan(&rec.Name, &rec.Value); err != nil {
					return err
				}
				records[rec.Name] = rec
			}

			return nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("internal/server/repositories/counter.go: query row context error %w", err)
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
