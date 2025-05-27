package agent

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

// Adapter defines the interface for making integration with server
type Adapter interface {
	UpdateMetric(ctx context.Context, mType, mName, value string) error
	UpdateMetrics(ctx context.Context, data []*models.Metric) error
}

type resetMetric func(ctx context.Context, a *Agent) error

// Agent handles the collection and reporting of system metrics
type Agent struct {
	lg                   *logging.ZapLogger
	cfg                  config.Config
	reporter             Adapter
	runtimeMetrics       []RuntimeMetric
	customMetrics        []*CustomMetric
	virtualMemoryMetrics []VirtualMemoryMetric
	cpuMetrics           []CPUMetric
	resetMetrics         []resetMetric
	reporterPipeLock     sync.Mutex
	repository           *MetricsRepository
	batchReport          bool
}

// NewAgent creates a new Agent instance with the specified configuration
func NewAgent(lg *logging.ZapLogger, cfg config.Config, rep *MetricsRepository, adaper Adapter) *Agent {
	agent := &Agent{
		lg:                   lg,
		cfg:                  cfg,
		reporter:             adaper,
		runtimeMetrics:       runtimeMetricsDefinition,
		customMetrics:        customMetricsDefinition,
		virtualMemoryMetrics: virtualMemoryMetricsDefinition,
		cpuMetrics:           cpuMetricsDefinition,
		reporterPipeLock:     sync.Mutex{},
		repository:           rep,
		batchReport:          cfg.BatchReport,
	}
	initResetMetrics(agent)

	return agent
}

// Start launches multiple goroutines:
// - startPoller: collects metrics
// - startReporter: sends metrics to the server
func (a *Agent) Start(ctx context.Context) {
	wg := sync.WaitGroup{}

	ctx = a.lg.WithContextFields(ctx, zap.String("name", "agent"))
	a.lg.InfoCtx(ctx, "init", zap.Any("config", a.cfg))
	a.startProfiler()
	a.startPoller(ctx, &wg)
	a.startReporter(ctx, &wg)
	wg.Wait()
}

// startProfiler starts the HTTP profiler server if configured
func (a *Agent) startProfiler() {
	if a.cfg.ProfileAddress == "" {
		return
	}

	go func() {
		if err := http.ListenAndServe(a.cfg.ProfileAddress, nil); err != nil {
			panic(err)
		}
	}()
}

// startPoller launches a goroutine that periodically collects system metrics
func (a *Agent) startPoller(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer func() {
			wg.Done()
			a.lg.InfoCtx(ctx, "poller pipe done")
		}()

		pollerCtx := a.lg.WithContextFields(ctx, zap.String("actor", "poller"))
		for {
			select {
			case <-pollerCtx.Done():
				a.lg.InfoCtx(pollerCtx, "poller done with context cancellation, do report")
				a.runReporterPipe(ctx)
				return
			default:
				a.lg.DebugCtx(ctx, "poller operation started")
				if err := a.runPollerPipe(pollerCtx); err != nil {
					a.lg.ErrorCtx(pollerCtx, "error in poller pipe", zap.Error(err))
				}

				a.lg.DebugCtx(ctx, "poller operation started")
				time.Sleep(a.cfg.PollInterval)
			}
		}
	}()
}

// startReporter launches a goroutine that periodically sends collected metrics to the server
func (a *Agent) startReporter(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		reporterCtx := a.lg.WithContextFields(ctx, zap.String("actor", "reporter"))

		defer func() {
			wg.Done()
			a.lg.InfoCtx(ctx, "reporter done")
		}()

		for {
			select {
			case <-reporterCtx.Done():
				a.lg.InfoCtx(reporterCtx, "reporter done with context cancellation")
				return
			case <-time.NewTicker(a.cfg.ReportInterval).C:
				a.lg.DebugCtx(reporterCtx, "reporter start")
				a.runReporterPipe(reporterCtx)
				a.lg.DebugCtx(reporterCtx, "sleep", zap.Duration("dur", a.cfg.ReportInterval))
			}
		}
	}()
}

// convertToStr converts various numeric types to their string representation
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
