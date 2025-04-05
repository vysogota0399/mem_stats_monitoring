package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/handlers/mocks"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap/zapcore"
)

func TestNewUpdatesRestMetricsHandler(t *testing.T) {
	type args struct {
		s storage.Storage
	}

	type fields struct {
		srvc  *mocks.MockIUpdateMetricsService
		input []byte
	}

	tests := []struct {
		name string
		args args
		fields
		prepare    func(f *fields)
		wantStatus int
	}{
		{
			name: "when invalid json input returns 400",
			args: args{
				s: storage.NewMemory(),
			},
			fields: fields{
				input: []byte(`{"invalid": "field"}`),
			},
			wantStatus: http.StatusBadRequest,
			prepare:    func(f *fields) {},
		},
		{
			name: "when service failed returns 500",
			args: args{
				s: storage.NewMemory(),
			},
			fields: fields{
				input: []byte(`[{"id": "id", "type": "gauge"}]`),
			},
			wantStatus: http.StatusInternalServerError,
			prepare: func(f *fields) {
				f.srvc.EXPECT().Call(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
			},
		},
		{
			name: "when service succeded returns 200",
			args: args{
				s: storage.NewMemory(),
			},
			fields: fields{
				input: []byte(`[{"id": "id", "type": "gauge"}]`),
			},
			wantStatus: http.StatusOK,
			prepare: func(f *fields) {
				f.srvc.EXPECT().Call(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
	}
	for _, tt := range tests {
		lg, err := logging.MustZapLogger(zapcore.DebugLevel)
		assert.NoError(t, err)

		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			fields := tt.fields
			fields.srvc = mocks.NewMockIUpdateMetricsService(ctrl)
			tt.prepare(&fields)

			h := NewUpdatesRestMetricsHandler(tt.args.s, fields.srvc, lg)
			router := gin.Default()
			route := "/updates/"
			router.POST(route, h)

			r, err := http.NewRequestWithContext(
				context.TODO(),
				http.MethodPost,
				route,
				bytes.NewBuffer(tt.input),
			)

			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			defer func() {
				if err = r.Body.Close(); err != nil {
					t.Errorf("failed to close request body: %v", err)
				}
			}()

			result := w.Result()
			defer func() {
				if err = result.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tt.wantStatus, result.StatusCode)
		})
	}
}
