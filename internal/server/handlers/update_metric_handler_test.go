package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

func TestNewUpdateMetricHandler(t *testing.T) {
	type want struct {
		statusCode int
	}

	tasks := []struct {
		name           string
		url            string
		method         string
		headers        map[string]string
		want           want
		metricsUpdater metricsUpdater
	}{
		{
			name:           "when valid request got response status ok",
			url:            "/update/gauge/TotalAlloc/0",
			method:         http.MethodPost,
			headers:        map[string]string{"Content-Type": "text/plain"},
			metricsUpdater: func(m Metric, storage storage.Storage) error { return nil },
			want:           want{statusCode: http.StatusOK},
		},
		{
			name:           "when invalid method response got status not found",
			url:            "/update/gauge/TotalAlloc/0",
			method:         http.MethodGet,
			headers:        map[string]string{"Content-Type": "text/plain"},
			metricsUpdater: func(m Metric, storage storage.Storage) error { return nil },
			want:           want{statusCode: http.StatusNotFound},
		},
		{
			name:           "when invalid url response got status not found",
			url:            "/update/0",
			method:         http.MethodGet,
			headers:        map[string]string{"Content-Type": "text/plain"},
			metricsUpdater: func(m Metric, storage storage.Storage) error { return nil },
			want:           want{statusCode: http.StatusNotFound},
		},
		{
			name:           "when invalid params response got status bad request",
			url:            "/update/hist/TotalAlloc/0",
			method:         http.MethodPost,
			headers:        map[string]string{"Content-Type": "text/plain"},
			metricsUpdater: func(m Metric, storage storage.Storage) error { return errors.New("error") },
			want:           want{statusCode: http.StatusBadRequest},
		},
	}

	for _, tt := range tasks {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			handler := updateMetricHandlerFunc(
				&UpdateMetricHandler{
					storage:        storage.New(),
					metricsUpdater: tt.metricsUpdater,
				},
			)
			router.POST("/update/:type/:name/:value", handler)

			r, err := http.NewRequestWithContext(context.TODO(), tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			if err != nil {
				t.Error(err)
			}

			for key, value := range tt.headers {
				r.Header.Add(key, value)
			}

			router.ServeHTTP(w, r)
			response := w.Result()
			assert.Equal(t, tt.want.statusCode, response.StatusCode, "%s %s \n%v", tt.method, tt.url, tt.headers)
			response.Body.Close()
		})
	}
}
