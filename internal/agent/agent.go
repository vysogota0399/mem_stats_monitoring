package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/clients"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type httpClient interface {
	UpdateMetric(ctx context.Context, mType, mName, value string) error
}

type Agent struct {
	ctx           context.Context
	lg            *logging.ZapLogger
	storage       storage.Storage
	cfg           config.Config
	httpClient    httpClient
	memoryMetics  []MemMetric
	customMetrics []CustomMetric
}

func NewAgent(ctx context.Context, lg *logging.ZapLogger, cfg config.Config, store storage.Storage) *Agent {
	return &Agent{
		ctx:           lg.WithContextFields(ctx, zap.String("name", "agent")),
		lg:            lg,
		storage:       store,
		cfg:           cfg,
		httpClient:    clients.NewCompReporter(ctx, cfg.ServerURL, lg),
		memoryMetics:  memMetricsDefinition,
		customMetrics: customMetricsDefinition,
	}
}
func (a Agent) Start() {
	a.lg.InfoCtx(a.ctx, "init", zap.Any("config", a.cfg))
	wg := sync.WaitGroup{}
	wg.Add(2)
	a.startPoller(&wg)
	a.startReporter(&wg)
	wg.Wait()
}

func (a *Agent) startPoller(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()

		ctx := a.lg.WithContextFields(a.ctx, zap.String("actor", "poller"))
		for {
			a.PollIteration(ctx)
			a.lg.DebugCtx(ctx, "sleep", zap.Duration("dur", a.cfg.PollInterval))
			time.Sleep(a.cfg.PollInterval)
		}
	}()
}

func (a Agent) startReporter(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()

		ctx := a.lg.WithContextFields(a.ctx, zap.String("actor", "reporter"))
		for {
			a.ReportIteration(ctx)
			a.lg.DebugCtx(ctx, "sleep", zap.Duration("dur", a.cfg.PollInterval))
			time.Sleep(a.cfg.ReportInterval)
		}
	}()
}

func (a *Agent) PollIteration(ctx context.Context) {
	operationID := uuid.NewV4()
	ctx = a.lg.WithContextFields(ctx, zap.String("operation_id", operationID.String()))
	a.lg.DebugCtx(ctx, "start")
	a.processMemMetrics(ctx)
	a.processCustomMetrics(ctx)
	a.lg.DebugCtx(ctx, "finished")
}

func (a Agent) ReportIteration(ctx context.Context) int {
	var counter int
	operationID := uuid.NewV4()
	ctx = a.lg.WithContextFields(ctx, zap.String("operation_id", operationID.String()))
	a.lg.DebugCtx(ctx, "start")

	for _, m := range a.memoryMetics {
		count, err := a.doReport(ctx, m)
		if err != nil && !errors.Is(err, storage.ErrNoRecords) {
			a.lg.ErrorCtx(ctx, "report mem metrics failed", zap.Error(err))
			continue
		}

		counter += count
	}

	for _, m := range a.customMetrics {
		count, err := a.doReport(ctx, m)
		if err != nil && !errors.Is(err, storage.ErrNoRecords) {
			a.lg.ErrorCtx(ctx, "report custom metrics failed", zap.Error(err))
			continue
		}

		counter += count
	}

	a.lg.DebugCtx(ctx, "finished")
	return counter
}

func (a *Agent) doReport(ctx context.Context, m Reportable) (int, error) {
	record, err := m.fromStore(a.storage)
	if err != nil {
		return 0, err
	}

	if err := a.httpClient.UpdateMetric(ctx, record.Type, record.Name, record.Value); err != nil {
		return 0, err
	}

	return 1, nil
}

func convertToStr(val any) (string, error) {
	switch val2 := val.(type) {
	case uint32:
		return fmt.Sprintf("%d", val2), nil
	case uint64:
		return fmt.Sprintf("%d", val2), nil
	case float64:
		return fmt.Sprintf("%.2f", val2), nil
	default:
		return "", fmt.Errorf("internal/agent: value %v underfined type - %T error", val, val)
	}
}
