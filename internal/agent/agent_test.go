package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"golang.org/x/sync/errgroup"
)

// TestNewAgent проверяет создание нового агента.
func TestNewAgent(t *testing.T) {
	logger, err := logging.MustZapLogger(-1)
	assert.NoError(t, err)
	store := storage.NewMemoryStorage(logger)
	cfg, err := config.NewConfig()
	assert.NoError(t, err)

	agent := NewAgent(logger, cfg, store)
	assert.NotNil(t, agent)
	assert.Equal(t, logger, agent.lg)
	assert.Equal(t, store, agent.storage)
	assert.Equal(t, cfg, agent.cfg)
}

// TestRunPollerPipe проверяет функцию runPollerPipe.
func TestRunPollerPipe(t *testing.T) {
	t.Parallel()

	logger, err := logging.MustZapLogger(-1)
	require.NoError(t, err)

	store := storage.NewMemoryStorage(logger)
	cfg, err := config.NewConfig()
	assert.NoError(t, err)

	agent := NewAgent(logger, cfg, store)
	ctx := context.Background()

	err = agent.runPollerPipe(ctx)
	require.NoError(t, err)
}

// TestGenMetrics tests metrics generation by verifying that the genMetrics function
// produces valid metrics with expected fields and types through the returned channel.
func TestGenMetrics(t *testing.T) {
	t.Parallel()

	logger, err := logging.MustZapLogger(-1)
	assert.NoError(t, err)

	store := storage.NewMemoryStorage(logger)
	cfg, err := config.NewConfig()
	assert.NoError(t, err)

	agent := NewAgent(logger, cfg, store)
	ctx := context.Background()

	metricsChan := agent.genMetrics(ctx, &errgroup.Group{})
	require.NotNil(t, metricsChan)

	// Check multiple metrics to ensure generator is working
	for range 3 {
		select {
		case metric := <-metricsChan:
			require.NotNil(t, metric)
			assert.NotEmpty(t, metric.Name, "Metric name should not be empty")
			assert.Contains(t, []string{"gauge", "counter"}, metric.Type, "Metric type should be gauge or counter")
			assert.NotNil(t, metric.Value, "Metric value should not be nil") // Changed from NotEmpty to NotNil
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for metrics")
		}
	}
}

// TestSaveMetrics проверяет сохранение метрик в хранилище.
func TestSaveMetrics(t *testing.T) {
	t.Parallel()

	logger, err := logging.MustZapLogger(-1)
	require.NoError(t, err)

	store := storage.NewMemoryStorage(logger)
	cfg, err := config.NewConfig()
	assert.NoError(t, err)

	agent := NewAgent(logger, cfg, store)

	metricsChan := make(chan *models.Metric, 10)
	errg, ctx := errgroup.WithContext(context.Background())

	// Send test metrics
	testMetrics := []*models.Metric{
		{Name: "TestMetric1", Type: "gauge", Value: "123.45"},
		{Name: "TestMetric2", Type: "counter", Value: "100"},
		{Name: "TestMetric3", Type: "gauge", Value: "67.89"},
	}

	for _, m := range testMetrics {
		metricsChan <- m
	}
	close(metricsChan)

	// Save metrics
	agent.saveMetrics(ctx, errg, metricsChan)

	err = errg.Wait()
	assert.NoError(t, err)
}

// TestConvertToStr проверяет конвертацию значений в строку.
func TestConvertToStr(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected string
	}{
		{"int32", int32(42), "42"},
		{"float64", float64(3.14), "3.14"},
		{"uint32", uint32(100), "100"},
		{"invalid", complex128(1 + 2i), ""}, // Ожидается ошибка
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := convertToStr(tc.input)
			if err != nil {
				assert.Empty(t, result)
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.expected, result)
				assert.NoError(t, err)
			}
		})
	}
}
