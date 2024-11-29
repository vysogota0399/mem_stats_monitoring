package agent

import (
	"fmt"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/clients"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

type httpClient interface {
	UpdateMetric(mType, mName, value string, requestID uuid.UUID) error
}

type Agent struct {
	logger        utils.Logger
	storage       storage.Storage
	config        Config
	httpClient    httpClient
	memoryMetics  []MemMetric
	customMetrics []CustomMetric
}

func NewAgent(config Config, store storage.Storage) *Agent {
	return &Agent{
		logger:        utils.InitLogger("[agent]"),
		storage:       store,
		config:        config,
		httpClient:    clients.NewReporter(config.ServerURL),
		memoryMetics:  memMetricsDefinition,
		customMetrics: customMetricsDefinition,
	}
}
func (a Agent) Start() {
	a.logger.Println(a.config)
	wg := sync.WaitGroup{}
	wg.Add(2)
	a.startPoller(&wg)
	a.startReporter(&wg)
	wg.Wait()
}

func (a *Agent) startPoller(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()

		for {
			a.PollIteration()
			a.logger.Printf("[poller] Sleep %v", a.config.PollInterval)
			time.Sleep(a.config.PollInterval)
		}
	}()
}

func (a Agent) startReporter(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()

		for {
			a.ReportIteration()
			a.logger.Printf("[reporter]Sleep %v", a.config.ReportInterval)
			time.Sleep(a.config.ReportInterval)
		}
	}()
}

func (a *Agent) PollIteration() {
	operationID := uuid.NewV4()
	a.logger.Printf("[poller][%v] OPERATION START", operationID)
	a.processMemMetrics(operationID)
	a.processCustomMetrics(operationID)
	a.logger.Printf("[poller][%v] OPERATION FINISHED", operationID)
}

func (a Agent) ReportIteration() int {
	var counter int
	operationID := uuid.NewV4()
	a.logger.Printf("[reporter][%v] OPERATION START", operationID)

	for _, m := range a.memoryMetics {
		count, err := a.doReport(m, operationID)
		if err != nil {
			a.logger.Printf("[reporter] err: %w", err)
			continue
		}

		counter += count
	}

	for _, m := range a.customMetrics {
		count, err := a.doReport(m, operationID)
		if err != nil {
			a.logger.Printf("[reporter] err: %w", err)
			continue
		}

		counter += count
	}

	a.logger.Printf("[reporter][%v] OPERATION FINISHED", operationID)
	return counter
}

func (a *Agent) doReport(m Reportable, operationID uuid.UUID) (int, error) {
	record, err := m.fromStore(a.storage)
	if err != nil {
		return 0, err
	}

	if err := a.httpClient.UpdateMetric(record.Type, record.Name, record.Value, operationID); err != nil {
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
		return "", fmt.Errorf("internal/agent: value %v underfined type - %t error", val, val)
	}
}
