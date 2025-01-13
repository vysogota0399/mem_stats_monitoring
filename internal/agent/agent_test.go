package agent

import (
	"context"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap/zapcore"
)

type mockClient struct{}

func (c *mockClient) UpdateMetric(ctx context.Context, mType, mName, value string) error {
	return nil
}

func (c *mockClient) UpdateMetrics(ctx context.Context, b []*models.Metric) error {
	return nil
}
func TestPollIteration(t *testing.T) {
	cfg, err := config.NewConfig()
	assert.NoError(t, err)

	lg, err := logging.MustZapLogger(zapcore.DebugLevel)
	assert.NoError(t, err)

	ctx := context.Background()

	agent := NewAgent(
		lg,
		cfg,
		storage.NewMemoryStorage(lg),
	)
	agent.httpClient = &mockClient{}

	agent.memoryMetics = []MemMetric{
		{
			Name: "Alloc", Type: "gauge",
			generateValue: func(stat *runtime.MemStats) any { return uint64(0) },
		},
		{
			Name: "UnexpectedType", Type: "gauge",
			generateValue: func(stat *runtime.MemStats) any { return 0 },
		},
	}

	agent.PollIteration(ctx)
	createdRecord, err := agent.storage.Get("gauge", "Alloc")
	assert.NoError(t, err)
	assert.Equal(t, &models.Metric{Name: "Alloc", Type: "gauge", Value: "0"}, createdRecord)

	skippedRecord, err := agent.storage.Get("gauge", "UnexpectedType")
	assert.Error(t, err)
	assert.Nil(t, skippedRecord)

	counterValue, err := agent.storage.Get("counter", "PollCount")
	assert.NoError(t, err)
	assert.Equal(t, &models.Metric{Name: "PollCount", Type: "counter", Value: "1"}, counterValue)

	randomValue, err := agent.storage.Get("gauge", "RandomValue")
	assert.NoError(t, err)
	assert.NotNil(t, randomValue.Value)
}

func TestReportIteration(t *testing.T) {
	cfg, err := config.NewConfig()
	assert.NoError(t, err)

	lg, err := logging.MustZapLogger(zapcore.DebugLevel)
	assert.NoError(t, err)

	ctx := context.Background()

	agent := NewAgent(
		lg,
		cfg,
		storage.NewMemoryStorage(lg),
	)
	agent.httpClient = &mockClient{}
	agent.memoryMetics = []MemMetric{}

	agent.PollIteration(ctx)
	result := agent.ReportIteration(ctx)
	assert.Equal(t, len(agent.customMetrics), result)
}
