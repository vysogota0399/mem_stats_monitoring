package agent

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"strconv"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
)

type MemValueGenerator func(*runtime.MemStats) any

type Reportable interface {
	fromStore(s storage.Storage) (*models.Metric, error)
}

type MemMetric struct {
	Name          string
	Type          string
	generateValue MemValueGenerator
}

func (m MemMetric) fromStore(s storage.Storage) (*models.Metric, error) {
	return s.Get(m.Type, m.Name)
}

type CustomMetric struct {
	Name          string
	Type          string
	generateValue func(mName, mType string, a *Agent) (uint64, error)
}

func (c CustomMetric) fromStore(s storage.Storage) (*models.Metric, error) {
	return s.Get(c.Type, c.Name)
}

var memMetricsDefinition []MemMetric = []MemMetric{
	{
		Name: "Alloc", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.Alloc },
	},
	{
		Name: "BuckHashSys", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.BuckHashSys },
	},
	{
		Name: "Frees", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.Frees },
	},
	{
		Name: "GCCPUFraction", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.GCCPUFraction },
	},
	{
		Name: "GCSys", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.GCSys },
	},
	{
		Name: "HeapAlloc", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.HeapAlloc },
	},
	{
		Name: "HeapIdle", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.HeapIdle },
	},
	{
		Name: "HeapInuse", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.HeapInuse },
	},
	{
		Name: "HeapObjects", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.HeapObjects },
	},
	{
		Name: "HeapReleased", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.HeapReleased },
	},
	{
		Name: "HeapSys", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.HeapSys },
	},
	{
		Name: "LastGC", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.LastGC },
	},
	{
		Name: "Lookups", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.Lookups },
	},
	{
		Name: "MCacheInuse", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.MCacheInuse },
	},
	{
		Name: "MCacheSys", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.MCacheSys },
	},
	{
		Name: "Mallocs", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.Mallocs },
	},
	{
		Name: "NextGC", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.NextGC },
	},
	{
		Name: "NumForcedGC", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.NumForcedGC },
	},
	{
		Name: "NumGC", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.NumGC },
	},
	{
		Name: "OtherSys", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.OtherSys },
	},
	{
		Name: "PauseTotalNs", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.PauseTotalNs },
	},
	{
		Name: "StackInuse", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.StackInuse },
	},
	{
		Name: "StackSys", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.StackSys },
	},
	{
		Name: "Sys", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.Sys },
	},
	{
		Name: "TotalAlloc", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.TotalAlloc },
	},
	{
		Name: "TotalAlloc", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.TotalAlloc },
	},
	{
		Name: "TotalAlloc", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.TotalAlloc },
	},
	{
		Name: "MSpanInuse", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.MSpanInuse },
	},
	{
		Name: "MSpanSys", Type: "gauge",
		generateValue: func(stat *runtime.MemStats) any { return stat.MSpanSys },
	},
}

var customMetricsDefinition = []CustomMetric{
	{
		Name: "PollCount",
		Type: "counter",
		generateValue: func(mName, mType string, a *Agent) (uint64, error) {
			var val uint64
			last, err := a.storage.Get(mType, mName)
			if err != nil && !errors.Is(err, storage.ErrNoRecords) {
				return val, err
			}

			if last == nil {
				val = 0
			} else {
				val, err = strconv.ParseUint(last.Value, 10, 64)
				if err != nil {
					return val, err
				}
			}

			return val + 1, nil
		},
	},
	{
		Name: "RandomValue",
		Type: "gauge",
		generateValue: func(mName, mType string, a *Agent) (uint64, error) {
			const max int64 = 100
			val, err := rand.Int(rand.Reader, big.NewInt(max))
			if err != nil {
				return 0, err
			}

			return val.Uint64(), nil
		},
	},
}

func (a *Agent) processMemMetrics(operationID uuid.UUID) {
	memStat := runtime.MemStats{}
	runtime.ReadMemStats(&memStat)

	for _, m := range a.memoryMetics {
		val, err := convertToStr(m.generateValue(&memStat))
		if err != nil {
			a.logger.Printf("[poller][%v] %v", operationID, err)
			continue
		}
		record := models.Metric{
			Name:  m.Name,
			Type:  m.Type,
			Value: val,
		}

		a.logger.Printf("[poller][%v] New value %s", operationID, record)
		if err = a.storage.Set(&record); err != nil {
			a.logger.Printf("[poller][%v] save value to storage error:%wvalue: %s", operationID, err, record)
			continue
		}
	}
}

func (a *Agent) processCustomMetrics(operationID uuid.UUID) {
	for _, m := range a.customMetrics {
		val, err := m.generateValue(m.Name, m.Type, a)
		if err != nil {
			a.logger.Printf("[poller][%v] generate value error:%w", operationID, err)
			continue
		}

		record := models.Metric{
			Name:  m.Name,
			Type:  m.Type,
			Value: fmt.Sprintf("%v", val),
		}

		a.logger.Printf("[poller][%v] New value %s", operationID, record)
		if err = a.storage.Set(&record); err != nil {
			a.logger.Printf("[poller][%v] save value to storage error:%wvalue: %s", operationID, err, record)
			continue
		}
	}
}
