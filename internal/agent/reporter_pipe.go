package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (a *Agent) runReporterPipe(ctx context.Context) {
	operationID := uuid.NewV4()
	ctx = a.lg.WithContextFields(ctx, zap.String("operation_id", operationID.String()))
	a.lg.InfoCtx(ctx, "start")

	g, ctx := errgroup.WithContext(ctx)

	a.report(ctx, g, a.loadMetrics(ctx, g))

	if err := g.Wait(); err != nil {
		a.lg.ErrorCtx(ctx, "report failed", zap.Error(err))
	}

	a.lg.InfoCtx(ctx, "finished")
}

func (a *Agent) loadMetrics(ctx context.Context, g *errgroup.Group) chan *models.Metric {
	metrics := make(chan *models.Metric)

	wg := sync.WaitGroup{}

	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()

		for _, m := range a.runtimeMetrics {
			if err := a.loadWithCancel(ctx, m, metrics); err != nil {
				return err
			}
		}

		return nil
	})

	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()

		for _, m := range a.customMetrics {
			if err := a.loadWithCancel(ctx, m, metrics); err != nil {
				return err
			}
		}

		return nil
	})

	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()

		for _, m := range a.virtualMemoryMetrics {
			if err := a.loadWithCancel(ctx, m, metrics); err != nil {
				return err
			}
		}

		return nil
	})

	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()

		for _, m := range a.cpuMetrics {
			if err := a.loadWithCancel(ctx, m, metrics); err != nil {
				return err
			}
		}

		return nil
	})

	go func() {
		wg.Wait()
		close(metrics)
	}()

	return metrics
}

func (a *Agent) loadWithCancel(
	ctx context.Context,
	r Reportable,
	b chan *models.Metric,
) error {
	m, err := r.fromStore(a.storage)
	if err != nil && !errors.Is(err, storage.ErrNoRecords) {
		return fmt.Errorf("internal/agent/reporter_pipe load from storage error %w", err)
	}

	select {
	case <-ctx.Done():
	case b <- m:
	}

	return nil
}

func (a *Agent) report(
	ctx context.Context,
	g *errgroup.Group,
	metrics chan *models.Metric,
) {
	g.Go(func() error {
		batch := make([]*models.Metric, 0)

		for m := range metrics {
			select {
			case <-ctx.Done():
				return nil
			default:
				g.Go(
					func() error {
						return a.httpClient.UpdateMetric(ctx, m.Type, m.Name, m.Value)
					},
				)
				batch = append(batch, m)
			}
		}

		if len(batch) == 0 {
			return nil
		}

		g.Go(
			func() error {
				return a.httpClient.UpdateMetrics(ctx, batch)
			},
		)

		return nil
	})
}
