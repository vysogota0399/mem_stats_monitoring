package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

// ICounterRepository interface for CounterRepository, it is used to mock CounterRepository in tests.
type ICounterRepository interface {
	Create(ctx context.Context, cntr *models.Counter) error
	FindByName(ctx context.Context, mName string) (models.Counter, error)
	All(ctx context.Context) ([]models.Counter, error)
	SaveCollection(ctx context.Context, coll []*models.Counter) error
	SearchByName(ctx context.Context, names []string) ([]models.Counter, error)
}

// CounterRepository отвечает за связь уровня бизнес логики и persistence layer в контексте работы с Counter.
type CounterRepository struct {
	storage storages.Storage
	lg      *logging.ZapLogger
}

func NewCounterRepository(strg storages.Storage, lg *logging.ZapLogger) *CounterRepository {
	return &CounterRepository{
		storage: strg,
		lg:      lg,
	}
}

// Create сохраняет новую запись в хранилище.
func (rep *CounterRepository) Create(ctx context.Context, cntr *models.Counter) error {
	record := models.Counter{
		Name: cntr.Name,
	}

	if err := rep.storage.GetCounter(ctx, &record); err != nil {
		if !errors.Is(err, storages.ErrNoRecords) {
			return fmt.Errorf("counter_repository.go: Create get counter error %w", err)
		}
	}

	cntr.Value += record.Value

	if err := rep.storage.CreateOrUpdate(ctx, models.CounterType, cntr.Name, cntr.Value); err != nil {
		return fmt.Errorf("counter_repository.go: create or update counter error %w", err)
	}

	return nil
}

// FindByName возвращает запись из хранилища по имени.
func (rep *CounterRepository) FindByName(ctx context.Context, mName string) (models.Counter, error) {
	record := models.Counter{
		Name: mName,
	}

	if err := rep.storage.GetCounter(ctx, &record); err != nil {
		return record, fmt.Errorf("counter_repository.go: FindByName get counter error %w", err)
	}

	return record, nil
}

// All возвращает все записи из хранилища.
func (rep *CounterRepository) All(ctx context.Context) ([]models.Counter, error) {
	records, err := rep.storage.GetCounters(ctx)
	if err != nil {
		return nil, fmt.Errorf("counter_repository.go: get counters error %w", err)
	}

	return records, nil
}

// SaveCollection сохраняет пачку записей в хранилище
func (rep *CounterRepository) SaveCollection(ctx context.Context, coll []models.Counter) error {
	operations := make([]func(ctx context.Context) error, 0, len(coll))
	repCtx := rep.lg.WithContextFields(ctx, zap.Any("actor", "counter_repository"))

	for _, cntr := range coll {
		operations = append(operations, func(repCtx context.Context) error {
			record := models.Counter{
				Name: cntr.Name,
			}

			if err := rep.storage.GetCounter(repCtx, &record); err != nil {
				if !errors.Is(err, storages.ErrNoRecords) {
					return fmt.Errorf("counter_repository.go: SaveCollection get counter error %+v %w", record, err)
				}
			}

			cntr.Value += record.Value

			return rep.storage.CreateOrUpdate(repCtx, models.CounterType, cntr.Name, cntr.Value)
		})
	}

	return rep.storage.Tx(repCtx, operations...)
}

// SearchByName осуществляет поиск записей по имени.
func (rep *CounterRepository) SearchByName(ctx context.Context, names []string) ([]models.Counter, error) {
	operations := make([]func(ctx context.Context) error, 0, len(names))
	records := make([]models.Counter, 0, len(names))

	for _, name := range names {
		operations = append(operations, func(ctx context.Context) error {
			cntr := models.Counter{
				Name: name,
			}

			if err := rep.storage.GetCounter(ctx, &cntr); err != nil {
				return fmt.Errorf("counter_repository.go: SearchByName get counter error %w", err)
			}

			records = append(records, cntr)
			return nil
		})
	}

	if err := rep.storage.Tx(ctx, operations...); err != nil {
		return nil, fmt.Errorf("counter_repository.go: tx error %w", err)
	}

	return records, nil
}
