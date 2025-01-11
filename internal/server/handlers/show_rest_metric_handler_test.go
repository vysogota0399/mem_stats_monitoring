package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap/zapcore"
)

func Test_showRestMetricHandlerFunc(t *testing.T) {
	const expectedRoute = "/value/"
	const expectedMType = "gauge"
	const expectedMName = "test"
	const unexpectedMName = "notfound"
	const expectedMValue = 1.0
	s := storage.NewMemory()
	record := models.Gauge{
		Name:  expectedMName,
		Value: expectedMValue,
	}
	rep := repositories.NewGauge(s)
	_, err := rep.Craete(context.Background(), &record)
	if err != nil {
		t.Fatal(err.Error())
	}

	type args struct {
		gaugeRepository   repositories.Gauge
		counterRepository repositories.Counter
		payload           []byte
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
				gaugeRepository:   repositories.NewGauge(s),
				counterRepository: repositories.NewCounter(s),
				payload:           []byte(fmt.Sprintf(`{"id": "%s", "type": "%s"}`, expectedMName, expectedMType)),
			},
			want: want{
				status:  200,
				payload: []byte(fmt.Sprintf(`{"id": "%s", "type": "%s", "value": %f}`, expectedMName, expectedMType, expectedMValue)),
				ctype:   "application/json",
			},
		},
		{
			name: "when invalid payload",
			args: args{
				gaugeRepository:   repositories.NewGauge(s),
				counterRepository: repositories.NewCounter(s),
				payload:           []byte(`{}`),
			},
			want: want{
				status:  400,
				payload: []byte(`{}`),
				ctype:   "application/json",
			},
		},
		{
			name: "when valid payload and record not found",
			args: args{
				gaugeRepository:   repositories.NewGauge(s),
				counterRepository: repositories.NewCounter(s),
				payload:           []byte(fmt.Sprintf(`{"id": "%s", "type": "%s"}`, unexpectedMName, expectedMType)),
			},
			want: want{
				status:  404,
				payload: []byte(`{}`),
				ctype:   "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			lg, err := logging.MustZapLogger(zapcore.DebugLevel)
			assert.NoError(t, err)

			handler := showRestMetricHandlerFunc(
				&ShowRestMetricHandler{
					gaugeRepository:   tt.args.gaugeRepository,
					counterRepository: tt.args.counterRepository,
					lg:                lg,
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
