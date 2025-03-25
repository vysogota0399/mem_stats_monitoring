package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap/zapcore"
)

func BenchmarkReporter_UpdateMetrics(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	client := NewMockRequester(ctrl)
	response := &http.Response{StatusCode: http.StatusOK}
	response.Body = io.NopCloser(bytes.NewBuffer([]byte{}))

	lg, err := logging.MustZapLogger(zapcore.ErrorLevel)
	assert.NoError(b, err)

	reporter := NewCompReporter("", lg, &config.Config{RateLimit: 10}, client)
	ctx := context.Background()

	types := []string{models.CounterType, models.CounterType}
	mCount := 10_000
	metrics := make([]*models.Metric, mCount)

	for b.Loop() {
		for i := range mCount {
			metrics[i] = &models.Metric{
				Name:  gofakeit.Animal(),
				Type:  types[rand.Int31n(2)],
				Value: gofakeit.Digit(),
			}
		}

		client.EXPECT().Request(gomock.Any()).Return(response, nil)
		assert.NoError(b, reporter.UpdateMetrics(ctx, metrics))
	}
}

func BenchmarkReporter_bytesreader(b *testing.B) {
	metrics := []*models.Metric{
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
		{Name: "asd", Type: "gauge", Value: "123"},
	}
	var body bytes.Buffer

	if err := json.NewEncoder(&body).Encode(metrics); err != nil {
		assert.NoError(b, err)
	}

	b.Run("when copy", func(b *testing.B) {
		for b.Loop() {
			buff := bytes.Buffer{}
			_, err := io.Copy(&buff, &body)
			assert.NoError(b, err)
		}
	})

	b.Run("when read all", func(b *testing.B) {
		for b.Loop() {
			_, err := io.ReadAll(&body)
			assert.NoError(b, err)
		}
	})
}
