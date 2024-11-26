package agent

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
)

type MemValueGenerator func(*runtime.MemStats) any

type MemMetric struct {
	Name          string
	Type          string
	generateValue MemValueGenerator
}

var memMetrics []MemMetric = []MemMetric{
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
}

func processMemMetrics(a *Agent, operationID uuid.UUID) {
	memStat := runtime.MemStats{}

	for _, m := range memMetrics {
		val, err := convertToStr(m.generateValue(&memStat))
		if err != nil {
			a.pollerLogger.Printf("[%v] %v", operationID, err)
			continue
		}
		record := models.Metric{
			Name:  m.Name,
			Type:  m.Type,
			Value: val,
		}

		a.pollerLogger.Printf("[%v] New value %s", operationID, record)
		if err = a.storage.Set(&record); err != nil {
			a.pollerLogger.Printf("[%v] save value to storage error\berror:%wvalue: %s", operationID, err, record)
			continue
		}
	}
}

func processCustomMetrics(a *Agent, operationID uuid.UUID) {
	metrics := []struct {
		Name          string
		Type          string
		generateValue func(mName, mType string, a *Agent) (uint64, error)
	}{
		{
			Name: "PollCount",
			Type: "counter",
			generateValue: func(mName, mType string, a *Agent) (uint64, error) {
				var val uint64
				last, err := a.storage.Get(mType, mName)
				if err != nil && err != storage.ErrNoRecords {
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
				return rand.Uint64(), nil
			},
		},
	}

	for _, m := range metrics {
		val, err := m.generateValue(m.Name, m.Type, a)
		if err != nil {
			a.pollerLogger.Printf("[%v] generate value error\berror:%w", operationID, err)
			continue
		}

		record := models.Metric{
			Name:  m.Name,
			Type:  m.Type,
			Value: fmt.Sprintf("%v", val),
		}

		a.pollerLogger.Printf("[%v] New value %s", operationID, record)
		if err = a.storage.Set(&record); err != nil {
			a.pollerLogger.Printf("[%v] save value to storage error\berror:%wvalue: %s", operationID, err, record)
			continue
		}
	}
}
