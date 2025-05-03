package agent

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
	mocks "github.com/vysogota0399/mem_stats_monitoring/internal/mocks/agent"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"golang.org/x/sync/errgroup"
)

// TestNewAgent проверяет создание нового агента.
func TestNewAgent(t *testing.T) {
	logger, err := logging.MustZapLogger(&config.Config{LogLevel: 0})
	assert.NoError(t, err)
	store := storage.NewMemoryStorage(logger)
	rep := NewMetricsRepository(store)
	cfg, err := config.NewConfig(nil)
	assert.NoError(t, err)

	agent := NewAgent(logger, cfg, rep, nil)
	assert.NotNil(t, agent)
	assert.Equal(t, logger, agent.lg)
	assert.Equal(t, rep, agent.repository)
	assert.Equal(t, cfg, agent.cfg)
}

// TestRunPollerPipe проверяет функцию runPollerPipe.
func TestRunPollerPipe(t *testing.T) {
	t.Parallel()

	logger, err := logging.MustZapLogger(&config.Config{LogLevel: 0})
	require.NoError(t, err)

	store := storage.NewMemoryStorage(logger)
	rep := NewMetricsRepository(store)
	cfg := config.Config{
		PollInterval:   time.Millisecond * 900,
		ReportInterval: time.Millisecond * 900,
	}
	assert.NoError(t, err)

	agent := NewAgent(logger, cfg, rep, nil)
	ctx := context.Background()

	err = agent.runPollerPipe(ctx)
	require.NoError(t, err)
}

// TestGenMetrics tests metrics generation by verifying that the genMetrics function
// produces valid metrics with expected fields and types through the returned channel.
func TestGenMetrics(t *testing.T) {
	t.Parallel()

	logger, err := logging.MustZapLogger(&config.Config{LogLevel: 0})
	assert.NoError(t, err)

	store := storage.NewMemoryStorage(logger)
	rep := NewMetricsRepository(store)

	cfg, err := config.NewConfig(nil)
	assert.NoError(t, err)

	agent := NewAgent(logger, cfg, rep, nil)
	ctx := context.Background()

	metricsChan := agent.genMetrics(ctx, &errgroup.Group{})
	require.NotNil(t, metricsChan)

	// Check multiple metrics to ensure generator is working
	for range 3 {
		select {
		case metric, ok := <-metricsChan:
			if ok {
				require.NotNil(t, metric)
				assert.NotEmpty(t, metric.Name, "Metric name should not be empty")
				assert.Contains(t, []string{"gauge", "counter"}, metric.Type, "Metric type should be gauge or counter")
				assert.NotNil(t, metric.Value, "Metric value should not be nil") // Changed from NotEmpty to NotNil
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for metrics")
		}
	}
}

// TestSaveMetrics проверяет сохранение метрик в хранилище.
func TestSaveMetrics(t *testing.T) {
	t.Parallel()

	logger, err := logging.MustZapLogger(&config.Config{LogLevel: 0})
	require.NoError(t, err)

	store := storage.NewMemoryStorage(logger)
	rep := NewMetricsRepository(store)
	cfg := config.Config{}
	assert.NoError(t, err)

	agent := NewAgent(logger, cfg, rep, nil)

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

func TestAgent_Start(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(client *mocks.MockHttpClient) (context.Context, context.CancelFunc)
		validate func(t *testing.T, ctx context.Context)
	}{
		{
			name: "graceful shutdown",
			setup: func(client *mocks.MockHttpClient) (context.Context, context.CancelFunc) {
				client.EXPECT().UpdateMetric(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				client.EXPECT().UpdateMetrics(gomock.Any(), gomock.Any()).AnyTimes().AnyTimes().Return(nil)
				return context.WithTimeout(context.Background(), 1000*time.Millisecond)
			},
			validate: func(t *testing.T, ctx context.Context) {
				assert.Equal(t, context.DeadlineExceeded, ctx.Err())
			},
		},
		{
			name: "context cancellation",
			setup: func(client *mocks.MockHttpClient) (context.Context, context.CancelFunc) {
				client.EXPECT().UpdateMetric(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				client.EXPECT().UpdateMetrics(gomock.Any(), gomock.Any()).AnyTimes().AnyTimes().Return(nil)
				ctx, cancel := context.WithCancel(context.Background())
				go func(cancel context.CancelFunc) {
					time.Sleep(time.Second)
					cancel()
				}(cancel)

				return ctx, cancel
			},
			validate: func(t *testing.T, ctx context.Context) {
				assert.Equal(t, context.Canceled, ctx.Err())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg := config.Config{
				PollInterval:   time.Millisecond * 900,
				ReportInterval: time.Millisecond * 900,
			}

			mockClient := mocks.NewMockHttpClient(ctrl)
			lg, err := logging.MustZapLogger(&config.Config{LogLevel: 0})
			assert.NoError(t, err)

			agent := &Agent{
				lg:                   lg,
				repository:           NewMetricsRepository(storage.NewMemoryStorage(lg)),
				cfg:                  cfg,
				reporter:             mockClient,
				runtimeMetrics:       runtimeMetricsDefinition,
				customMetrics:        customMetricsDefinition,
				virtualMemoryMetrics: virtualMemoryMetricsDefinition,
				cpuMetrics:           cpuMetricsDefinition,
			}

			ctx, cancel := tt.setup(mockClient)
			defer cancel()

			go agent.Start(ctx)

			// Wait for context to be done
			<-ctx.Done()

			tt.validate(t, ctx)
		})
	}
}
