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

// Gauge отвечает за связь уровня бизнес логики и persistence layer в контексте работы с Gauge.
type Gauge struct {
	storage storage.Storage
	Records []models.Gauge
	lg      *logging.ZapLogger
}

func NewGauge(strg storage.Storage, lg *logging.ZapLogger) *Gauge {
	return &Gauge{
		storage: strg,
		Records: make([]models.Gauge, 0),
		lg:      lg,
	}
}

// Create сохраняет новую запись в хранилище.
func (g *Gauge) Create(ctx context.Context, record *models.Gauge) (*models.Gauge, error) {
	if s, ok := g.storage.(DBAble); ok {
		return g.pushToDB(ctx, s, record)
	}

	if err := g.storage.Push(models.GaugeType, record.Name, record); err != nil {
		return record, err
	}

	return record, nil
}

func (g *Gauge) pushToDB(ctx context.Context, s DBAble, r *models.Gauge) (*models.Gauge, error) {
	if err := s.QueryRowContext(
		ctx,
		`
			INSERT INTO gauges(name, value)
			VALUES ($1, $2)
			RETURNING id
		`,
		[]any{r.Name, r.Value},
		func(rows *sql.Rows) error {
			return rows.Scan(&r.ID)
		},
	); err != nil {
		return nil, err
	}

	return r, nil
}

// Last возвращает последнюю запись из хранилища.
func (g Gauge) Last(ctx context.Context, mName string) (*models.Gauge, error) {
	if s, ok := g.storage.(DBAble); ok {
		return g.lastFromDB(ctx, s, mName)
	}

	return g.lastFromMem(mName)
}

func (g Gauge) lastFromDB(ctx context.Context, s DBAble, mName string) (*models.Gauge, error) {
	gg := &models.Gauge{Name: mName}

	if err := s.QueryRowContext(
		ctx,
		`
		SELECT value, id
		FROM gauges
		WHERE name = $1
		ORDER BY id DESC
		LIMIT 1`,
		[]any{mName},
		func(rows *sql.Rows) error {
			return rows.Scan(&gg.Value, &gg.ID)
		},
	); err != nil {
		return nil, fmt.Errorf("internal/server/repositories/gauge.go: query row context error %w", err)
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

// All возвращает все записи из хранилища.
func (g Gauge) All() map[string][]models.Gauge { //nolint:dupl // :/
	records := map[string][]models.Gauge{}
	mNames, ok := g.storage.All()[models.GaugeType]
	if !ok {
		return records
	}

	for name, values := range mNames {
		count := len(values)
		collection := make([]models.Gauge, count)
		for i := range count {
			collection[i] = models.Gauge{}
			if err := json.Unmarshal([]byte(values[i]), &collection[i]); err != nil {
				continue
			}
		}
		records[name] = collection
	}

	return records
}

// SaveCollection сохраняет пачку записей в хранилище
func (g *Gauge) SaveCollection(ctx context.Context, coll []models.Gauge) ([]models.Gauge, error) {
	if s, ok := g.storage.(DBAble); ok {
		return g.saveCollToDB(ctx, s, coll)
	}

	return g.saveCollToMem(coll)
}

func (g *Gauge) saveCollToDB(ctx context.Context, s DBAble, coll []models.Gauge) ([]models.Gauge, error) {
	tx, err := s.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	defer func() {
		if commitErr := s.CommitTx(ctx, tx); commitErr != nil {
			if rollbackErr := s.RollbackTx(ctx, tx); rollbackErr != nil {
				g.lg.ErrorCtx(ctx, "failed to rollback transaction after commit error", zap.Error(rollbackErr))
			}
		}
	}()

	for _, rec := range coll {
		if err := s.QueryRowContext(
			ctx,
			`
			INSERT INTO gauges(name, value)
			VALUES ($1, $2)
			RETURNING id
			`,
			[]any{rec.Name, rec.Value},
			func(rows *sql.Rows) error {
				return rows.Scan(&rec.ID)
			},
		); err != nil {
			if rollbackErr := s.RollbackTx(ctx, tx); rollbackErr != nil {
				g.lg.ErrorCtx(ctx, "failed to rollback transaction after scan error", zap.Error(rollbackErr))
			}
			return nil, fmt.Errorf("internal/server/repositories/gauge.go: query row context error %w", err)
		}
	}

	return coll, nil
}

func (g *Gauge) saveCollToMem(coll []models.Gauge) ([]models.Gauge, error) {
	for _, rec := range coll {
		if err := g.storage.Push(models.GaugeType, rec.Name, &rec); err != nil {
			return nil, err
		}
	}

	return coll, nil
}
