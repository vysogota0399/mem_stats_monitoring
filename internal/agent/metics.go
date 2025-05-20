package agent

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"strconv"
	"sync"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
)

// MemValueGenerator is a function type that generates metric values from runtime memory statistics
type MemValueGenerator func(*runtime.MemStats) any

// Reportable interface defines the contract for metrics that can be loaded from storage
type Reportable interface {
	Load(rep *MetricsRepository) (*models.Metric, error)
}

// RuntimeMetric represents a metric that can be collected from runtime memory statistics
type RuntimeMetric struct {
	Name          string
	Type          string
	generateValue MemValueGenerator
	Reset         resetMetric
}

// Load loads the metric value from storage
func (m RuntimeMetric) Load(rep *MetricsRepository) (*models.Metric, error) {
	return rep.Get(m.Name, m.Type)
}

// runtimeMetricsDefinition defines the set of runtime memory metrics to collect
var runtimeMetricsDefinition = []RuntimeMetric{
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

type CustomMetric struct {
	Name          string
	Type          string
	lock          sync.Mutex
	generateValue func(*CustomMetric, *Agent) (uint64, error)
	Reset         resetMetric
}

func (c *CustomMetric) Load(rep *MetricsRepository) (*models.Metric, error) {
	return rep.Get(c.Name, c.Type)
}

var customMetricsDefinition = []*CustomMetric{
	{
		Name: "PollCount",
		Type: models.CounterType,
		lock: sync.Mutex{},
		generateValue: func(m *CustomMetric, a *Agent) (uint64, error) {
			var pollCount uint64
			var err error

			m.lock.Lock()
			defer m.lock.Unlock()

			metric, err := a.repository.Get(m.Name, m.Type)
			// ошибка отличается от "не найдено"
			if err != nil {
				if !errors.Is(err, storage.ErrNoRecords) {
					return pollCount, err
				}

				metric = a.repository.New(m.Name, m.Type, "0")
			}
			defer a.repository.Release(metric)

			_, _, value := a.repository.SafeRead(metric)
			pollCount, err = strconv.ParseUint(value, 10, 64)
			if err != nil {
				return pollCount, fmt.Errorf("customMetricsDefinition: parse string %s error %w", value, err)
			}

			pollCount++

			return pollCount, nil
		},
		Reset: func(ctx context.Context, a *Agent) error {
			metric, err := a.repository.Get("PollCount", models.CounterType)
			if err != nil {
				return fmt.Errorf("customMetricsDefinition: reset metric failed error: %w", err)
			}
			metric.Value = "0"

			return a.repository.SaveAndRelease(ctx, metric)
		},
	},
	{
		Name: "RandomValue",
		Type: "gauge",
		generateValue: func(m *CustomMetric, a *Agent) (uint64, error) {
			const max int64 = 100
			val, err := rand.Int(rand.Reader, big.NewInt(max))
			if err != nil {
				return 0, err
			}

			return val.Uint64(), nil
		},
	},
}

type VirtualMemoryMetric struct {
	Name          string
	Type          string
	generateValue func(*mem.VirtualMemoryStat) uint64
	Reset         resetMetric
}

func (c VirtualMemoryMetric) Load(rep *MetricsRepository) (*models.Metric, error) {
	return rep.Get(c.Name, c.Type)
}

var virtualMemoryMetricsDefinition = []VirtualMemoryMetric{
	{
		Name: "TotalMemory", Type: "gauge",
		generateValue: func(stat *mem.VirtualMemoryStat) uint64 { return stat.Total },
	},
	{
		Name: "FreeMemory", Type: "gauge",
		generateValue: func(stat *mem.VirtualMemoryStat) uint64 { return stat.Free },
	},
}

type CPUMetric struct {
	Name          string
	Type          string
	generateValue func([]cpu.InfoStat) int32
	Reset         resetMetric
}

func (c CPUMetric) Load(rep *MetricsRepository) (*models.Metric, error) {
	return rep.Get(c.Name, c.Type)
}

var cpuMetricsDefinition = []CPUMetric{
	{
		Name: "CPUutilization1", Type: "gauge",
		generateValue: func(stat []cpu.InfoStat) int32 {
			var sum int32
			for _, c := range stat {
				sum += c.CPU
			}

			return sum
		},
	},
}

func initResetMetrics(a *Agent) {
	a.resetMetrics = make([]resetMetric, 0)

	for _, m := range a.runtimeMetrics {
		if m.Reset != nil {
			a.resetMetrics = append(a.resetMetrics, m.Reset)
		}
	}

	for _, m := range a.customMetrics {
		if m.Reset != nil {
			a.resetMetrics = append(a.resetMetrics, m.Reset)
		}
	}

	for _, m := range a.virtualMemoryMetrics {
		if m.Reset != nil {
			a.resetMetrics = append(a.resetMetrics, m.Reset)
		}
	}

	for _, m := range a.cpuMetrics {
		if m.Reset != nil {
			a.resetMetrics = append(a.resetMetrics, m.Reset)
		}
	}
}
