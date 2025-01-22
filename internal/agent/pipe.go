package agent

import (
	"context"
	"runtime"
	"sync"

	uuid "github.com/satori/go.uuid"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"go.uber.org/zap"
)

func (a *Agent) runPipe(ctx context.Context) {
	operationID := uuid.NewV4()
	ctx = a.lg.WithContextFields(ctx, zap.String("operation_id", operationID.String()))
	a.lg.DebugCtx(ctx, "start")

	a.saveMetrics(ctx, a.genMetrics(ctx))

	a.lg.DebugCtx(ctx, "finished")
}

func (a *Agent) saveMetrics(ctx context.Context, metrics <-chan *models.Metric) {
	numWorkers := 10
	wg := sync.WaitGroup{}
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		func() {
			defer wg.Done()

			for m := range metrics {
				select {
				case <-ctx.Done():
					return
				default:
					if err := a.storage.Set(ctx, m); err != nil {
						a.lg.ErrorCtx(ctx, "save to storate error", zap.Error(err), zap.Any("metric", m))
					}
				}
			}
		}()
	}

	wg.Wait()
}

func (a *Agent) genMetrics(ctx context.Context) chan *models.Metric {
	metrics := make(chan *models.Metric)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		memStat := &runtime.MemStats{}
		runtime.ReadMemStats(memStat)

		for _, m := range a.runtimeMetrics {
			val, err := convertToStr(m.generateValue(memStat))
			if err != nil {
				a.lg.ErrorCtx(ctx, "convert to str error", zap.Error(err), zap.Any("metric", m))
				// TODO: ERROR
			}
			metrics <- &models.Metric{
				Name:  m.Name,
				Type:  m.Type,
				Value: val,
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, m := range a.customMetrics {
			val, err := m.generateValue(&m, a)
			if err != nil {
				a.lg.ErrorCtx(ctx, "generate val error", zap.Error(err), zap.Any("metric", m))
				// TODO: ERROR
			}

			sVal, err := convertToStr(val)
			if err != nil {
				a.lg.ErrorCtx(ctx, "convert to str error", zap.Error(err), zap.Any("metric", m))
				// TODO: ERROR
			}
			metrics <- &models.Metric{
				Name:  m.Name,
				Type:  m.Type,
				Value: sVal,
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		stat, err := mem.VirtualMemory()
		if err != nil {
			a.lg.ErrorCtx(ctx, "calc virtual memory error", zap.Error(err))
			// TODO: ERROR
		}

		for _, m := range a.virtualMemoryMetrics {
			val, err := convertToStr(m.generateValue(stat))
			if err != nil {
				a.lg.ErrorCtx(ctx, "convert to str error", zap.Error(err), zap.Any("metric", m))
				// TODO: ERROR
			}

			metrics <- &models.Metric{
				Name:  m.Name,
				Type:  m.Type,
				Value: val,
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		stats, err := cpu.Info()
		if err != nil {
			a.lg.ErrorCtx(ctx, "calc virtual memory error", zap.Error(err))
			// TODO: ERROR
		}

		for _, m := range a.cpuMetrics {
			val, err := convertToStr(m.generateValue(stats))
			if err != nil {
				a.lg.ErrorCtx(ctx, "convert to str error", zap.Error(err), zap.Any("metric", m))
				// TODO: ERROR
			}

			metrics <- &models.Metric{
				Name:  m.Name,
				Type:  m.Type,
				Value: val,
			}
		}
	}()

	go func() {
		wg.Wait()
		close(metrics)
	}()

	return metrics
}
