package agent

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func BenchmarkAgent_genCustromMetrics(b *testing.B) {
	type fields struct {
		lg            *logging.ZapLogger
		storage       *storage.Memory
		cfg           config.Config
		customMetrics []CustomMetric
	}

	lg, err := logging.MustZapLogger(zap.DebugLevel)
	assert.NoError(b, err)
	cfg := config.Config{}

	benchFields := fields{
		lg:            lg,
		storage:       storage.NewMemoryStorage(lg),
		cfg:           cfg,
		customMetrics: customMetricsDefinition,
	}
	ctx := context.Background()
	wg := sync.WaitGroup{}
	errg, ctx := errgroup.WithContext(ctx)
	agent := &Agent{
		lg:            benchFields.lg,
		cfg:           benchFields.cfg,
		storage:       benchFields.storage,
		customMetrics: benchFields.customMetrics,
		metricsPool:   NewMetricsPool(),
	}

	b.Run("fetch objects from pool", func(b *testing.B) {
		resChan := make(chan *models.Metric)
		go func(c chan *models.Metric) { <-c }(resChan)

		for i := 0; i < b.N; i++ {
			agent.genCustromMetrics(
				ctx,
				&wg,
				errg,
				resChan,
				true,
			)
		}
	})

	b.Run("allocates new object", func(b *testing.B) {
		resChan := make(chan *models.Metric)
		go func(c chan *models.Metric) {
			agent.metricsPool.Put(<-c)
		}(resChan)

		for i := 0; i < b.N; i++ {
			agent.genCustromMetrics(
				ctx,
				&wg,
				errg,
				resChan,
				true,
			)
		}
	})
}

func BenchmarkAgent_genRuntimeMetrics(b *testing.B) {
	type fields struct {
		lg             *logging.ZapLogger
		storage        *storage.Memory
		cfg            config.Config
		runtimeMetrics []RuntimeMetric
	}

	lg, err := logging.MustZapLogger(zap.DebugLevel)
	assert.NoError(b, err)
	cfg := config.Config{}

	benchFields := fields{
		lg:             lg,
		storage:        storage.NewMemoryStorage(lg),
		cfg:            cfg,
		runtimeMetrics: runtimeMetricsDefinition,
	}
	ctx := context.Background()
	wg := sync.WaitGroup{}
	errg, ctx := errgroup.WithContext(ctx)
	agent := &Agent{
		lg:             benchFields.lg,
		cfg:            benchFields.cfg,
		storage:        benchFields.storage,
		runtimeMetrics: benchFields.runtimeMetrics,
		metricsPool:    NewMetricsPool(),
	}

	b.Run("fetch objects from pool", func(b *testing.B) {
		resChan := make(chan *models.Metric)
		go func(c chan *models.Metric) {
			agent.metricsPool.Put(<-c)
		}(resChan)

		for b.Loop() {
			agent.genRuntimeMetrics(
				ctx,
				&wg,
				errg,
				resChan,
				true,
			)
		}
	})

	b.Run("allocates new object", func(b *testing.B) {
		resChan := make(chan *models.Metric)
		go func(c chan *models.Metric) { <-c }(resChan)

		for b.Loop() {
			agent.genRuntimeMetrics(
				ctx,
				&wg,
				errg,
				resChan,
				false,
			)
		}
	})
}

func BenchmarkAgent_genVirtualMemoryMetrics(b *testing.B) {
	type fields struct {
		lg            *logging.ZapLogger
		storage       *storage.Memory
		cfg           config.Config
		virtualMemory []VirtualMemoryMetric
	}

	lg, err := logging.MustZapLogger(zap.DebugLevel)
	assert.NoError(b, err)
	cfg := config.Config{}

	benchFields := fields{
		lg:            lg,
		storage:       storage.NewMemoryStorage(lg),
		cfg:           cfg,
		virtualMemory: virtualMemoryMetricsDefinition,
	}
	ctx := context.Background()
	wg := sync.WaitGroup{}
	errg, ctx := errgroup.WithContext(ctx)
	agent := &Agent{
		lg:                   benchFields.lg,
		cfg:                  benchFields.cfg,
		storage:              benchFields.storage,
		virtualMemoryMetrics: benchFields.virtualMemory,
		metricsPool:          NewMetricsPool(),
	}

	b.Run("fetch objects from pool", func(b *testing.B) {
		resChan := make(chan *models.Metric)
		go func(c chan *models.Metric) {
			agent.metricsPool.Put(<-c)
		}(resChan)

		for i := 0; i < b.N; i++ {
			agent.genVirtualMemoryMetrics(
				ctx,
				&wg,
				errg,
				resChan,
				true,
			)
		}
	})

	b.Run("allocate new object", func(b *testing.B) {
		resChan := make(chan *models.Metric)
		go func(c chan *models.Metric) { <-c }(resChan)

		for i := 0; i < b.N; i++ {
			agent.genVirtualMemoryMetrics(
				ctx,
				&wg,
				errg,
				resChan,
				false,
			)
		}
	})
}

func BenchmarkAgent_genCPUMetrics(b *testing.B) {
	type fields struct {
		lg         *logging.ZapLogger
		storage    *storage.Memory
		cfg        config.Config
		cpuMetrics []CPUMetric
	}

	lg, err := logging.MustZapLogger(zap.DebugLevel)
	assert.NoError(b, err)
	cfg := config.Config{}

	benchFields := fields{
		lg:         lg,
		storage:    storage.NewMemoryStorage(lg),
		cfg:        cfg,
		cpuMetrics: cpuMetricsDefinition,
	}
	ctx := context.Background()
	wg := sync.WaitGroup{}
	errg, ctx := errgroup.WithContext(ctx)
	agent := &Agent{
		lg:          benchFields.lg,
		cfg:         benchFields.cfg,
		storage:     benchFields.storage,
		cpuMetrics:  benchFields.cpuMetrics,
		metricsPool: NewMetricsPool(),
	}

	b.Run("fetch objects from pool", func(b *testing.B) {
		resChan := make(chan *models.Metric)
		go func(c chan *models.Metric) {
			agent.metricsPool.Put(<-c)
		}(resChan)

		for b.Loop() {
			agent.genCPUMetrics(
				ctx,
				&wg,
				errg,
				resChan,
				true,
			)
		}
	})

	b.Run("allocate new object", func(b *testing.B) {
		resChan := make(chan *models.Metric)
		go func(c chan *models.Metric) { <-c }(resChan)
		for b.Loop() {
			agent.genCPUMetrics(
				ctx,
				&wg,
				errg,
				resChan,
				false,
			)
		}
	})
}
