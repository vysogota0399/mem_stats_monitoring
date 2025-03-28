package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap/zapcore"
)

type succededSerice struct{}
type failureSerice struct{}

var expectedValue float64 = 1

func (s succededSerice) Call(ctx context.Context, p service.UpdateMetricServiceParams) (service.UpdateMetricServiceResult, error) {
	return service.UpdateMetricServiceResult{Value: &expectedValue, MType: models.GaugeType, ID: "test"}, nil
}

func (s failureSerice) Call(ctx context.Context, p service.UpdateMetricServiceParams) (service.UpdateMetricServiceResult, error) {
	return service.UpdateMetricServiceResult{}, errors.New("error")
}

func Test_updateRestMetricHandlerFunc(t *testing.T) {
	var expectedRoute = "/update/"

	type args struct {
		service UpdateMetricService
		payload []byte
	}
	type want struct {
		status  int
		payload []byte
		ctype   string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "when valid payload and service returns no errors",
			args: args{
				service: succededSerice{},
				payload: []byte(`{"id": "test", "type": "gauge", "value": 1}`),
			},
			want: want{
				status:  200,
				payload: []byte(fmt.Sprintf(`{"id": "test", "type": "gauge", "value": %v}`, expectedValue)),
				ctype:   "application/json",
			},
		},
		{
			name: "when invalid payload",
			args: args{
				service: succededSerice{},
				payload: []byte(`{"id": "test"}`),
			},
			want: want{
				status:  400,
				payload: []byte(`{}`),
				ctype:   "application/json",
			},
		},
		{
			name: "when valid payload and service failed",
			args: args{
				service: failureSerice{},
				payload: []byte(`{"id": "test"}`),
			},
			want: want{
				status:  400,
				payload: []byte(`{}`),
				ctype:   "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			s := storage.NewMemory()
			lg, err := logging.MustZapLogger(zapcore.DebugLevel)
			assert.NoError(t, err)

			handler := updateRestMetricHandlerFunc(
				&UpdateRestMetricHandler{
					storage: s,
					service: tt.args.service,
					lg:      lg,
				},
			)

			router.POST(expectedRoute, handler)

			r, err := http.NewRequestWithContext(
				context.TODO(),
				http.MethodPost,
				expectedRoute,
				bytes.NewBuffer(tt.args.payload),
			)

			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			defer r.Body.Close()

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.status, result.StatusCode)
			assert.JSONEq(t, string(tt.want.payload), w.Body.String())
			assert.Equal(t, tt.want.ctype, w.Header().Get("Content-Type"))
		})
	}
}
