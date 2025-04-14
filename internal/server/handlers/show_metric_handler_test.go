package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestNewShowMetricHandler(t *testing.T) {
	type want struct {
		statusCode int
		response   string
	}
	type args struct {
		storage storage.Storage
		mName   string
		mType   string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "when valid request got expected metric then status 200",
			want: want{
				response:   "1",
				statusCode: http.StatusOK,
			},
			args: args{
				storage: storage.NewMemStorageWithData(map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}}),
				mName:   "test",
				mType:   "counter",
			},
		},
		{
			name: "when invalid metric type then status 404",
			want: want{
				response:   "",
				statusCode: http.StatusNotFound,
			},
			args: args{
				storage: storage.NewMemStorageWithData(map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}}),
				mName:   "test",
				mType:   "hist",
			},
		},
		{
			name: "when metric not found then status 404",
			want: want{
				response:   "",
				statusCode: http.StatusNotFound,
			},
			args: args{
				storage: storage.NewMemStorageWithData(map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}}),
				mName:   "unexpectedName",
				mType:   "counter",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			handler := NewShowMetricHandler(tt.args.storage, lg)
			router.GET("/value/:type/:name", handler)

			r, err := http.NewRequestWithContext(context.TODO(), "GET", fmt.Sprintf("/value/%s/%s", tt.args.mType, tt.args.mName), nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			if err != nil {
				assert.NoError(t, err)
			}

			router.ServeHTTP(w, r)
			response := w.Result()
			defer func() {
				if err := response.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tt.want.statusCode, response.StatusCode)
			assert.Equal(t, tt.want.response, w.Body.String())
		})
	}
}
