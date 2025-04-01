package agent

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	uuid "github.com/satori/go.uuid"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (a *Agent) runPollerPipe(ctx context.Context) error {
	operationID := uuid.NewV4()
	ctx = a.lg.WithContextFields(ctx, zap.String("operation_id", operationID.String()))
	a.lg.DebugCtx(ctx, "start")

	g, ctx := errgroup.WithContext(ctx)

	a.saveMetrics(ctx, g, a.genMetrics(ctx, g))

	if err := g.Wait(); err != nil {
		return fmt.Errorf("poller_pile: collect metrics failed error %w", err)
	}

	a.lg.DebugCtx(ctx, "finished")
	return nil
}

func (a *Agent) saveMetrics(ctx context.Context, g *errgroup.Group, metrics <-chan *models.Metric) {
	numWorkers := 10

	for range numWorkers {
		g.Go(func() error {
			for m := range metrics {
				select {
				case <-ctx.Done():
					return nil
				default:
					if err := a.storage.Set(ctx, m); err != nil {
						a.lg.ErrorCtx(ctx, "save to storate error", zap.Error(err), zap.Any("metric", m))
						a.metricsPool.Put(m)
						return fmt.Errorf("internal/agent/poller_pipe save to storate error %w", err)
					} else {
						a.metricsPool.Put(m)
					}
				}
			}

			return nil
		})
	}
}

func (a *Agent) genMetrics(ctx context.Context, g *errgroup.Group) chan *models.Metric {
	wg := &sync.WaitGroup{}
	metrics := make(chan *models.Metric)
	done := make(chan struct{})

	// Start all metric generators
	a.genRuntimeMetrics(ctx, wg, g, metrics, done, true)
	a.genCustromMetrics(ctx, wg, g, metrics, done, true)
	a.genVirtualMemoryMetrics(ctx, wg, g, metrics, done, true)
	a.genCPUMetrics(ctx, wg, g, metrics, done, true)

	// Close metrics channel when all generators are done
	g.Go(func() error {
		wg.Wait()
		close(done)
		close(metrics)
		return nil
	})

	return metrics
}

func (a *Agent) genCPUMetrics(
	ctx context.Context,
	wg *sync.WaitGroup,
	g *errgroup.Group,
	metrics chan *models.Metric,
	done chan struct{},
	fromPool bool,
) {
	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()

		stats, err := cpu.Info()
		if err != nil {
			return fmt.Errorf("internal/agent/poller_pipe calc cpu error %w", err)
		}

		for _, m := range a.cpuMetrics {
			select {
			case <-ctx.Done():
				a.lg.InfoCtx(ctx, "genCPUMetrics context done with context cancellation")
				return nil
			case <-done:
				return nil
			default:
				val, err := convertToStr(m.generateValue(stats))
				if err != nil {
					return fmt.Errorf("internal/agent/poller_pipe convert to str error %w", err)
				}

				var res *models.Metric
				if fromPool {
					res = a.metricsPool.Get()
					res.Name = m.Name
					res.Type = m.Type
					res.Value = val
				} else {
					res = &models.Metric{
						Name:  m.Name,
						Type:  m.Type,
						Value: val,
					}
				}

				select {
				case metrics <- res:
				case <-ctx.Done():
					if fromPool {
						a.metricsPool.Put(res)
					}
					a.lg.InfoCtx(ctx, "genCPUMetrics context done with context cancellation")
					return nil
				case <-done:
					if fromPool {
						a.metricsPool.Put(res)
					}
					return nil
				}
			}
		}

		return nil
	})
}

func (a *Agent) genVirtualMemoryMetrics(
	ctx context.Context,
	wg *sync.WaitGroup,
	g *errgroup.Group,
	metrics chan *models.Metric,
	done chan struct{},
	fromPool bool,
) {
	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()

		stat, err := mem.VirtualMemory()
		if err != nil {
			return fmt.Errorf("internal/agent/poller_pipe calc virtual memory error %w", err)
		}

		for _, m := range a.virtualMemoryMetrics {
			select {
			case <-ctx.Done():
				a.lg.InfoCtx(ctx, "genVirtualMemoryMetrics context done with context cancellation")
				return nil
			case <-done:
				return nil
			default:
				val, err := convertToStr(m.generateValue(stat))
				if err != nil {
					return fmt.Errorf("internal/agent/poller_pipe convert to str error %w", err)
				}

				var res *models.Metric
				if fromPool {
					res = a.metricsPool.Get()
					res.Name = m.Name
					res.Type = m.Type
					res.Value = val
				} else {
					res = &models.Metric{
						Name:  m.Name,
						Type:  m.Type,
						Value: val,
					}
				}

				select {
				case metrics <- res:
				case <-ctx.Done():
					if fromPool {
						a.metricsPool.Put(res)
					}
					a.lg.InfoCtx(ctx, "genVirtualMemoryMetrics context done with context cancellation")
					return nil
				case <-done:
					if fromPool {
						a.metricsPool.Put(res)
					}
					return nil
				}
			}
		}

		return nil
	})
}

func (a *Agent) genCustromMetrics(
	ctx context.Context,
	wg *sync.WaitGroup,
	g *errgroup.Group,
	metrics chan *models.Metric,
	done chan struct{},
	fromPool bool,
) {
	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()

		for _, m := range a.customMetrics {
			select {
			case <-ctx.Done():
				a.lg.InfoCtx(ctx, "genCustromMetrics context done with context cancellation")
				return nil
			case <-done:
				return nil
			default:
				val, err := m.generateValue(&m, a)
				if err != nil {
					return fmt.Errorf("internal/agent/poller_pipe generate val error %w", err)
				}

				sVal, err := convertToStr(val)
				if err != nil {
					return fmt.Errorf("internal/agent/poller_pipe generate val error %w", err)
				}

				var res *models.Metric
				if fromPool {
					res = a.metricsPool.Get()
					res.Name = m.Name
					res.Type = m.Type
					res.Value = sVal
				} else {
					res = &models.Metric{
						Name:  m.Name,
						Type:  m.Type,
						Value: sVal,
					}
				}

				select {
				case metrics <- res:
				case <-ctx.Done():
					if fromPool {
						a.metricsPool.Put(res)
					}
					a.lg.InfoCtx(ctx, "genCustromMetrics context done with context cancellation")
					return nil
				case <-done:
					if fromPool {
						a.metricsPool.Put(res)
					}
					return nil
				}
			}
		}

		return nil
	})
}

func (a *Agent) genRuntimeMetrics(
	ctx context.Context,
	wg *sync.WaitGroup,
	g *errgroup.Group,
	metrics chan *models.Metric,
	done chan struct{},
	fromPool bool,
) {
	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()

		memStat := &runtime.MemStats{}
		runtime.ReadMemStats(memStat)

		for _, m := range a.runtimeMetrics {
			select {
			case <-ctx.Done():
				a.lg.InfoCtx(ctx, "genRuntimeMetrics context done with context cancellation")
				return nil
			case <-done:
				return nil
			default:
				val, err := convertToStr(m.generateValue(memStat))
				if err != nil {
					return fmt.Errorf("internal/agent/poller_pipe convert to str error %w", err)
				}

				var res *models.Metric
				if fromPool {
					res = a.metricsPool.Get()
					res.Name = m.Name
					res.Type = m.Type
					res.Value = val
				} else {
					res = &models.Metric{
						Name:  m.Name,
						Type:  m.Type,
						Value: val,
					}
				}

				select {
				case metrics <- res:
				case <-ctx.Done():
					if fromPool {
						a.metricsPool.Put(res)
					}
					a.lg.InfoCtx(ctx, "genRuntimeMetrics context done with context cancellation")
					return nil
				case <-done:
					if fromPool {
						a.metricsPool.Put(res)
					}
					return nil
				}
			}
		}

		return nil
	})
}
