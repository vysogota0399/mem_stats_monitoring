package storages

import (
	"context"
	"errors"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/fx"
)

type Storage interface {
	Tx(ctx context.Context, fns ...func(ctx context.Context) error) error
	CreateOrUpdate(ctx context.Context, mType, mName string, val any) error
	GetCounter(ctx context.Context, record *models.Counter) error
	GetGauge(ctx context.Context, record *models.Gauge) error
	GetCounters(ctx context.Context) ([]models.Counter, error)
	GetGauges(ctx context.Context) ([]models.Gauge, error)
	Ping(ctx context.Context) error
	IncrementCounter(ctx context.Context, name string, delta int64) error
}

var ErrNoRecords = errors.New("memory: no records")

func NewStorage(
	lc fx.Lifecycle,
	dumper Dumper,
	lg *logging.ZapLogger,
	cfg *config.Config,
	migrator Migrator,
	connectionOpener ConnectionOpener,
	srsb SourceBuilder,
) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		return NewPG(lc, lg, cfg, migrator, connectionOpener)
	}

	if cfg.FileStoragePath != "" {
		return NewPersistance(lc, cfg, dumper, lg, srsb)
	}

	return NewMemory(lg), nil
}
