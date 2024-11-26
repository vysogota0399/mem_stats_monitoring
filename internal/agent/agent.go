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

type Agent struct {
	pollerLogger   utils.Logger
	reporterLogger utils.Logger
	storage        storage.Storage
	config         Config
	httpClient     clients.MemStatsServer
}

func NewAgent() *Agent {
	config := NewConfig()
	return &Agent{
		pollerLogger:   utils.InitLogger("[poller]"),
		reporterLogger: utils.InitLogger("[reporter]"),
		storage:        storage.NewMemoryStorage(),
		config:         config,
		httpClient:     clients.NewMemStatsServer(config.ServerURL),
	}
}
func (a Agent) Start() {
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
			operationID := uuid.NewV4()
			a.pollerLogger.Printf("[%v] OPERATION START", operationID)
			processMemMetrics(a, operationID)
			processCustomMetrics(a, operationID)
			a.pollerLogger.Printf("[%v] OPERATION FINISHED", operationID)

			a.pollerLogger.Printf("Sleep %v", a.config.PollInterval)
			time.Sleep(a.config.PollInterval)
		}
	}()
}

func (a Agent) startReporter(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()

		for {
			operationID := uuid.NewV4()
			a.reporterLogger.Printf("[%v] OPERATION START", operationID)
			for _, m := range memMetrics {
				record, err := a.storage.Get(m.Type, m.Name)
				if err != nil {
					if err == storage.ErrNoRecords {
						a.reporterLogger.Println(err)
						continue
					} else {
						panic(err)
					}
				}

				if err := a.httpClient.UpdateMetric(record.Type, record.Name, record.Value, operationID); err != nil {
					a.reporterLogger.Println(err)
				}
			}

			a.reporterLogger.Printf("[%v] OPERATION FINISHED", operationID)
			a.reporterLogger.Printf("Sleep %v", a.config.ReportInterval)
			time.Sleep(a.config.ReportInterval)
		}
	}()
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
		return "", fmt.Errorf("agent: value %v underfined type - %t error", val, val)
	}
}
