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
	lg            *logging.ZapLogger
	storage       storage.Storage
	cfg           config.Config
	httpClient    httpClient
	memoryMetics  []MemMetric
	customMetrics []CustomMetric
}

func NewAgent(lg *logging.ZapLogger, cfg config.Config, store storage.Storage) *Agent {
	return &Agent{
		lg:            lg,
		storage:       store,
		cfg:           cfg,
		httpClient:    clients.NewCompReporter(cfg.ServerURL, lg),
		memoryMetics:  memMetricsDefinition,
		customMetrics: customMetricsDefinition,
	}
}
func (a *Agent) Start(ctx context.Context) {
	ctx = a.lg.WithContextFields(ctx, zap.String("name", "agent"))
	a.lg.InfoCtx(ctx, "init", zap.Any("config", a.cfg))
	wg := sync.WaitGroup{}
	a.startPoller(ctx, &wg)
	a.startReporter(ctx, &wg)
	a.startBatchReporter(ctx, &wg)
	wg.Wait()
}

func (a *Agent) startPoller(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		ctx := a.lg.WithContextFields(ctx, zap.String("actor", "poller"))
		for {
			a.PollIteration(ctx)
			a.lg.DebugCtx(ctx, "sleep", zap.Duration("dur", a.cfg.PollInterval))
			time.Sleep(a.cfg.PollInterval)
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
				a.ReportIteration(ctx)
				a.lg.DebugCtx(ctx, "sleep", zap.Duration("dur", a.cfg.PollInterval))
				time.Sleep(a.cfg.ReportInterval)
			}
		}
	}()
}

func (a *Agent) startBatchReporter(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		ctx := a.lg.WithContextFields(ctx, zap.String("actor", "batch_reporter"))
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.NewTicker(a.cfg.ReportInterval).C:
				a.ReportBatch(ctx)
				a.lg.DebugCtx(ctx, "sleep", zap.Duration("dur", a.cfg.PollInterval))
			}
		}
	}()
}

func (a *Agent) ReportBatch(ctx context.Context) {
	operationID := uuid.NewV4()
	ctx = a.lg.WithContextFields(ctx, zap.String("operation_id", operationID.String()))

	a.lg.DebugCtx(ctx, "start")

	batch := make([]*models.Metric, 0)
	for _, m := range a.memoryMetics {
		m, err := m.fromStore(a.storage)
		if err != nil && !errors.Is(err, storage.ErrNoRecords) {
			a.lg.ErrorCtx(ctx, "fetch metric failed error", zap.Error(err))
			continue
		}

		batch = append(batch, m)
	}
	for _, m := range a.customMetrics {
		m, err := m.fromStore(a.storage)
		if err != nil && !errors.Is(err, storage.ErrNoRecords) {
			a.lg.ErrorCtx(ctx, "fetch metric failed error", zap.Error(err))
			continue
		}
		batch = append(batch, m)
	}

	if len(batch) == 0 {
		a.lg.DebugCtx(ctx, "batch is empty")
	} else {
		if err := a.httpClient.UpdateMetrics(ctx, batch); err != nil {
			a.lg.ErrorCtx(ctx, "batch report failed error", zap.Error(err))
		}
	}

	a.lg.DebugCtx(ctx, "finished")
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
