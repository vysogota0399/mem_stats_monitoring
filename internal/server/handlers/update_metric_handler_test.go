package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
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
			metricsUpdater: func(m Metric, storage storage.Storage, logger utils.Logger) error { return nil },
			want:           want{statusCode: http.StatusOK},
		},
		{
			name:           "when invalid headers got response status bad request",
			url:            "/update/gauge/TotalAlloc/0",
			method:         http.MethodPost,
			headers:        map[string]string{},
			metricsUpdater: func(m Metric, storage storage.Storage, logger utils.Logger) error { return nil },
			want:           want{statusCode: http.StatusBadRequest},
		},
		{
			name:           "when invalid method response got status not found",
			url:            "/update/gauge/TotalAlloc/0",
			method:         http.MethodGet,
			headers:        map[string]string{"Content-Type": "text/plain"},
			metricsUpdater: func(m Metric, storage storage.Storage, logger utils.Logger) error { return nil },
			want:           want{statusCode: http.StatusNotFound},
		},
		{
			name:           "when invalid url response got status not found",
			url:            "/update/0",
			method:         http.MethodGet,
			headers:        map[string]string{"Content-Type": "text/plain"},
			metricsUpdater: func(m Metric, storage storage.Storage, logger utils.Logger) error { return nil },
			want:           want{statusCode: http.StatusNotFound},
		},
		{
			name:           "when invalid params response got status bad request",
			url:            "/update/hist/TotalAlloc/0",
			method:         http.MethodPost,
			headers:        map[string]string{"Content-Type": "text/plain"},
			metricsUpdater: func(m Metric, storage storage.Storage, logger utils.Logger) error { return errors.New("error") },
			want:           want{statusCode: http.StatusBadRequest},
		},
	}

	for _, tt := range tasks {
		t.Run(tt.name, func(t *testing.T) {
			r, err := http.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			if err != nil {
				t.Error(err)
			}

			for key, value := range tt.headers {
				r.Header.Add(key, value)
			}

			subject := UpdateMetricHandler{logger: utils.InitLogger("[test]"), storage: nil, metricsUpdater: tt.metricsUpdater}
			subject.ServeHTTP(w, r)

			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode, "%s %s \n%v", tt.method, tt.url, tt.headers)
		})
	}
}