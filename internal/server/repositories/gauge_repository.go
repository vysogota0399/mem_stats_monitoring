package repositories

import (
	"context"
	"fmt"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

// IGaugeRepository interface for GaugeRepository, it is used to mock GaugeRepository in tests.
type IGaugeRepository interface {
	Create(ctx context.Context, record *models.Gauge) error
	FindByName(ctx context.Context, mName string) error
	All(ctx context.Context) ([]models.Gauge, error)
	SaveCollection(ctx context.Context, coll []models.Gauge) error
}

// GaugeRepository отвечает за связь уровня бизнес логики и persistence layer в контексте работы с Gauge.
type GaugeRepository struct {
	storage storages.Storage
	lg      *logging.ZapLogger
}

func NewGaugeRepository(strg storages.Storage, lg *logging.ZapLogger) *GaugeRepository {
	return &GaugeRepository{
		storage: strg,
		lg:      lg,
	}
}

// Create сохраняет новую запись в хранилище.
func (g *GaugeRepository) Create(ctx context.Context, record *models.Gauge) error {
	if err := g.storage.CreateOrUpdate(ctx, models.GaugeType, record.Name, record.Value); err != nil {
		return fmt.Errorf("gauge_repository.go: create or update gauge error %w", err)
	}

	return nil
}

// FindByName возвращает запись из хранилища по имени.
func (g *GaugeRepository) FindByName(ctx context.Context, mName string) (models.Gauge, error) {
	record := models.Gauge{
		Name: mName,
	}

	if err := g.storage.GetGauge(ctx, &record); err != nil {
		return models.Gauge{}, fmt.Errorf("gauge_repository.go: get gauge error %w", err)
	}

	return record, nil
}

// All возвращает все записи из хранилища.
func (g *GaugeRepository) All(ctx context.Context) ([]models.Gauge, error) {
	records, err := g.storage.GetGauges(ctx)
	if err != nil {
		return nil, fmt.Errorf("gauge_repository.go: get gauges error %w", err)
	}

	return records, nil
}

// SaveCollection сохраняет пачку записей в хранилище
func (g *GaugeRepository) SaveCollection(ctx context.Context, coll []models.Gauge) error {
	repCtx := g.lg.WithContextFields(ctx, zap.Any("actor", "gauge_repository"))

	operations := make([]func(ctx context.Context) error, 0, len(coll))

	for _, rec := range coll {
		operations = append(operations, func(repCtx context.Context) error {
			return g.storage.CreateOrUpdate(repCtx, models.GaugeType, rec.Name, rec.Value)
		})
	}

	return g.storage.Tx(repCtx, operations...)
}
