package agent

import (
	"runtime"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
)

type mockClient struct{}

func (c *mockClient) UpdateMetric(mType, mName, value string, requestID uuid.UUID) error {
	return nil
}
func TestPollIteration(t *testing.T) {
	agent := NewAgent(
		NewConfig(
			10*time.Second,
			2*time.Second,
			"http://test.com",
		),
		storage.NewMemoryStorage(),
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

	agent.PollIteration()
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
	agent := NewAgent(
		NewConfig(
			10*time.Second,
			2*time.Second,
			"http://test.com",
		),
		storage.NewMemoryStorage(),
	)
	agent.httpClient = &mockClient{}
	agent.memoryMetics = []MemMetric{}

	agent.PollIteration()
	result := agent.ReportIteration()
	assert.Equal(t, len(agent.customMetrics), result)
}
