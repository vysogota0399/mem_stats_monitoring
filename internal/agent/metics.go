package agent

import (
	"crypto/rand"
	"errors"
	"math/big"
	"runtime"
	"strconv"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
)

type MemValueGenerator func(*runtime.MemStats) any

type Reportable interface {
	fromStore(s storage.Storage) (*models.Metric, error)
}

type RuntimeMetric struct {
	Name          string
	Type          string
	generateValue MemValueGenerator
}

func (m RuntimeMetric) fromStore(s storage.Storage) (*models.Metric, error) {
	return s.Get(m.Type, m.Name)
}

type CustomMetric struct {
	Name          string
	Type          string
	generateValue func(*CustomMetric, *Agent) (uint64, error)
}

func (c CustomMetric) fromStore(s storage.Storage) (*models.Metric, error) {
	return s.Get(c.Type, c.Name)
}

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

var customMetricsDefinition = []CustomMetric{
	{
		Name: "PollCount",
		Type: "counter",
		generateValue: func(m *CustomMetric, a *Agent) (uint64, error) {
			var val uint64
			last, err := a.storage.Get(m.Type, m.Name)
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
}

func (c VirtualMemoryMetric) fromStore(s storage.Storage) (*models.Metric, error) {
	return s.Get(c.Type, c.Name)
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
}

func (c CPUMetric) fromStore(s storage.Storage) (*models.Metric, error) {
	return s.Get(c.Type, c.Name)
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
