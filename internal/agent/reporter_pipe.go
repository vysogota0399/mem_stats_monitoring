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

// runReporterPipe executes the complete reporting pipeline:
// 1. Loads metrics from storage
// 2. Sends metrics to the server
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

// loadMetrics loads metrics from various sources in parallel using errgroup
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

// loadWithCancel loads a single metric from storage with context cancellation support
func (a *Agent) loadWithCancel(
	ctx context.Context,
	r Reportable,
	b chan *models.Metric,
) error {
	m := a.metricsPool.Get()
	if err := r.fromStore(a.storage, m); err != nil && !errors.Is(err, storage.ErrNoRecords) {
		a.metricsPool.Put(m)
		return fmt.Errorf("internal/agent/reporter_pipe load from storage error %w", err)
	}

	select {
	case <-ctx.Done():
	case b <- m:
	}

	return nil
}

// report sends metrics to the server in batches
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
						if err := a.httpClient.UpdateMetric(ctx, m.Type, m.Name, m.Value); err != nil {
							a.metricsPool.Put(m)
							return fmt.Errorf("report_pipe: upload metric err %w", err)
						}

						return nil
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
				if err := a.httpClient.UpdateMetrics(ctx, batch); err != nil {
					for _, m := range batch {
						a.metricsPool.Put(m)
					}

					return fmt.Errorf("reporter_pipe: update batch metrics failed error %w", err)
				}

				for _, m := range batch {
					a.metricsPool.Put(m)
				}
				return nil
			},
		)

		return nil
	})
}
