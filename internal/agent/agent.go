package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/clients"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type httpClient interface {
	UpdateMetric(ctx context.Context, mType, mName, value string) error
	UpdateMetrics(ctx context.Context, data []*models.Metric) error
}

type Agent struct {
	lg                   *logging.ZapLogger
	storage              storage.Storage
	cfg                  config.Config
	httpClient           httpClient
	runtimeMetrics       []RuntimeMetric
	customMetrics        []CustomMetric
	virtualMemoryMetrics []VirtualMemoryMetric
	cpuMetrics           []CPUMetric
}

func NewAgent(lg *logging.ZapLogger, cfg config.Config, store storage.Storage) *Agent {
	return &Agent{
		lg:                   lg,
		storage:              store,
		cfg:                  cfg,
		httpClient:           clients.NewCompReporter(cfg.ServerURL, lg, &cfg),
		runtimeMetrics:       runtimeMetricsDefinition,
		customMetrics:        customMetricsDefinition,
		virtualMemoryMetrics: virtualMemoryMetricsDefinition,
		cpuMetrics:           cpuMetricsDefinition,
	}
}
func (a *Agent) Start(ctx context.Context) {
	ctx = a.lg.WithContextFields(ctx, zap.String("name", "agent"))
	a.lg.InfoCtx(ctx, "init", zap.Any("config", a.cfg))
	wg := sync.WaitGroup{}
	a.startPoller(ctx, &wg)
	a.startReporter(ctx, &wg)
	wg.Wait()
}

func (a *Agent) startPoller(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		ctx := a.lg.WithContextFields(ctx, zap.String("actor", "poller"))
		for {
			select {
			case <-ctx.Done():
				return
			default:
				a.runPollerPipe(ctx)
				a.lg.DebugCtx(ctx, "sleep", zap.Duration("dur", a.cfg.PollInterval))
				time.Sleep(a.cfg.PollInterval)
			}
		}
	}()
}

func (a Agent) startReporter(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		ctx := a.lg.WithContextFields(ctx, zap.String("actor", "reporter"))
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.NewTicker(a.cfg.ReportInterval).C:
				a.runReporterPipe(ctx)
				a.lg.DebugCtx(ctx, "sleep", zap.Duration("dur", a.cfg.PollInterval))
				time.Sleep(a.cfg.ReportInterval)
			}
		}
	}()
}

func convertToStr(val any) (string, error) {
	switch val2 := val.(type) {
	case uint32:
		return fmt.Sprintf("%d", val2), nil
	case int32:
		return fmt.Sprintf("%d", val2), nil
	case uint64:
		return fmt.Sprintf("%d", val2), nil
	case float64:
		return fmt.Sprintf("%.2f", val2), nil
	default:
		return "", fmt.Errorf("internal/agent: value %v underfined type - %T error", val, val)
	}
}
