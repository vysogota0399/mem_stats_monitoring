package storages

import (
	"context"
	"errors"
	"fmt"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storages/dump"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Persistance struct {
	*Memory
	lg       *logging.ZapLogger
	dumper   Dumper
	restorer Restorer
}

var ErrUnexpectedReader = errors.New("got unexpected reader type")

type Dumper interface {
	Dump(ctx context.Context, m dump.DumpMessage)
	Start(ctx context.Context) error
	Stop(ctx context.Context)
}

type Restorer interface {
	Call(ctx context.Context) error
}

func NewPersistance(lc fx.Lifecycle, cfg *config.Config, dumper Dumper, lg *logging.ZapLogger) (*Persistance, error) {
	strg := &Persistance{
		Memory: NewMemory(lg),
		lg:     lg,
	}

	restorer, err := NewFileMetricsRestorer(cfg, lg, strg)
	if err != nil {
		return nil, fmt.Errorf("persistance: failed to create restorer %w", err)
	}

	strg.restorer = restorer

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				if err := strg.restore(ctx); err != nil {
					return fmt.Errorf("persistance: restore failed error %w", err)
				}

				strg.dumper = dumper
				go func() {
					if err := strg.dumper.Start(ctx); err != nil {
						lg.ErrorCtx(ctx, "persistance: start dumper error", zap.Error(err))
					}
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {
				strg.dumper.Stop(ctx)

				return nil
			},
		},
	)

	return strg, nil
}

func (p *Persistance) CreateOrUpdate(ctx context.Context, mType, mName string, val any) error {
	if err := p.Memory.CreateOrUpdate(ctx, mType, mName, val); err != nil {
		return err
	}

	if p.dumper == nil {
		p.lg.DebugCtx(ctx, "dumper is not set, skipping dump. May be it's restore time now")
		return nil
	}

	p.dumper.Dump(
		p.lg.WithContextFields(ctx, zap.String("action", "create_or_update")),
		dump.DumpMessage{
			MName:  mName,
			MType:  mType,
			MValue: val,
		},
	)

	return nil
}

func (p *Persistance) restore(ctx context.Context) error {
	if err := p.restorer.Call(
		p.lg.WithContextFields(ctx, zap.String("action", "restore")),
	); err != nil {
		return fmt.Errorf("persistance: restore failed error %w", err)
	}

	return nil
}
