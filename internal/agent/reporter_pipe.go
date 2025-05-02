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
	a.reporterPipeLock.Lock()
	defer a.reporterPipeLock.Unlock()

	operationID := uuid.NewV4()
	ctx = a.lg.WithContextFields(ctx, zap.String("operation_id", operationID.String()))

	g, ctx := errgroup.WithContext(ctx)

	a.report(ctx, g, a.loadMetrics(g))

	if err := g.Wait(); err != nil {
		a.lg.ErrorCtx(ctx, "report failed", zap.Error(err))
	}

	a.lg.InfoCtx(ctx, "finished")
}

// loadMetrics loads metrics from various sources in parallel using errgroup
func (a *Agent) loadMetrics(g *errgroup.Group) chan *models.Metric {
	metrics := make(chan *models.Metric)

	wg := sync.WaitGroup{}

	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()

		for _, m := range a.runtimeMetrics {
			if err := a.load(m, metrics); err != nil {
				return err
			}
		}

		return nil
	})

	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()

		for _, m := range a.customMetrics {
			if err := a.load(m, metrics); err != nil {
				return err
			}
		}

		return nil
	})

	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()

		for _, m := range a.virtualMemoryMetrics {
			if err := a.load(m, metrics); err != nil {
				return err
			}
		}

		return nil
	})

	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()

		for _, m := range a.cpuMetrics {
			if err := a.load(m, metrics); err != nil {
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

// load loads a single metric from storage
func (a *Agent) load(
	r Reportable,
	b chan *models.Metric,
) error {
	m := a.metricsPool.Get()
	if err := r.fromStore(a.storage, m); err != nil && !errors.Is(err, storage.ErrNoRecords) {
		a.metricsPool.Put(m)
		return fmt.Errorf("internal/agent/reporter_pipe load from storage error %w", err)
	}

	b <- m

	return nil
}

// report sends metrics to the server in batches
func (a *Agent) report(
	ctx context.Context,
	g *errgroup.Group,
	metrics chan *models.Metric,
) {
	batch := make([]*models.Metric, 0)
	batchLock := &sync.Mutex{}

	for m := range metrics {
		g.Go(
			func() error {
				if err := a.reporter.UpdateMetric(ctx, m.Type, m.Name, m.Value); err != nil {
					a.metricsPool.Put(m)
					return fmt.Errorf("report_pipe: upload metric err %w", err)
				}

				return nil
			},
		)

		batchLock.Lock()
		batch = append(batch, m)
		batchLock.Unlock()
	}

	if len(batch) == 0 {
		return
	}

	g.Go(
		func() error {
			defer a.metricsPool.Free(batch)
			if err := a.reporter.UpdateMetrics(ctx, batch); err != nil {
				return fmt.Errorf("reporter_pipe: update batch metrics failed error %w", err)
			}

			return nil
		},
	)
}
