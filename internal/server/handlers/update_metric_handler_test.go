package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/repositories"
	mock "github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestNewUpdateMetricHandler(t *testing.T) {
	type args struct {
		service *repositories.MockICounterRepository
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "creates instance of UpdateMetricHandler",
			args: args{},
		},
	}
	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.service = repositories.NewMockICounterRepository(ctrl)

			service := service.NewUpdateMetricService(
				tt.args.service,
				repositories.NewMockIGaugeRepository(ctrl),
			)
			h := NewUpdateMetricHandler(service, lg)
			assert.NotNil(t, h)
			assert.Equal(t, service, h.service)
			assert.Equal(t, lg, h.lg)
		})
	}
}

func TestUpdateMetricHandler_Registrate(t *testing.T) {
	type fields struct {
		service IUpdateMetricService
		lg      *logging.ZapLogger
	}
	tests := []struct {
		name   string
		fields fields
		want   server.Route
	}{
		{
			name: "when service is provided",
			fields: fields{
				service: service.NewUpdateMetricService(
					repositories.NewMockICounterRepository(nil),
					repositories.NewMockIGaugeRepository(nil),
				),
				lg: nil,
			},
			want: server.Route{
				Path:   "/update/:type/:name/:value",
				Method: "POST",
			},
		},
		{
			name: "when service is nil",
			fields: fields{
				service: nil,
				lg:      nil,
			},
			want: server.Route{
				Path:   "/update/:type/:name/:value",
				Method: "POST",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &UpdateMetricHandler{
				service: tt.fields.service,
				lg:      tt.fields.lg,
			}
			got, err := h.Registrate()
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Path, got.Path)
			assert.Equal(t, tt.want.Method, got.Method)
			assert.NotNil(t, got.Handler)
		})
	}
}

func TestUpdateMetricHandler_handler(t *testing.T) {
	type fields struct {
		srv   *mock.MockIUpdateMetricService
		route string
	}
	tests := []struct {
		name           string
		fields         fields
		prepare        func(*fields)
		wantStatusCode int
	}{
		{
			name:           "when invalid metric type",
			fields:         fields{route: "/update/invalid/test/1"},
			wantStatusCode: http.StatusBadRequest,
			prepare:        func(fields *fields) {},
		},
		{
			name:           "when empty metric name",
			fields:         fields{route: "/update/gauge//1"},
			wantStatusCode: http.StatusBadRequest,
			prepare:        func(fields *fields) {},
		},
		{
			name:           "when invalid gauge value",
			fields:         fields{route: "/update/gauge/test/invalid"},
			wantStatusCode: http.StatusBadRequest,
			prepare:        func(fields *fields) {},
		},
		{
			name:           "when invalid counter value",
			fields:         fields{route: "/update/counter/test/invalid"},
			wantStatusCode: http.StatusBadRequest,
			prepare:        func(fields *fields) {},
		},
		{
			name:   "when gauge update success",
			fields: fields{route: "/update/gauge/test/1.5"},
			prepare: func(fields *fields) {
				fields.srv.EXPECT().Call(gomock.Any(), gomock.Any()).Return(service.UpdateMetricServiceResult{
					ID:    "test",
					MType: models.GaugeType,
					Value: 1.5,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "when counter update success",
			fields: fields{route: "/update/counter/test/1"},
			prepare: func(fields *fields) {
				fields.srv.EXPECT().Call(gomock.Any(), gomock.Any()).Return(service.UpdateMetricServiceResult{
					ID:    "test",
					MType: models.CounterType,
					Delta: 1,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "when service error",
			fields: fields{route: "/update/gauge/test/1.5"},
			prepare: func(fields *fields) {
				fields.srv.EXPECT().Call(gomock.Any(), gomock.Any()).Return(service.UpdateMetricServiceResult{}, errors.New("service error"))
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.srv = mock.NewMockIUpdateMetricService(ctrl)
			tt.prepare(&tt.fields)

			h := NewUpdateMetricHandler(tt.fields.srv, lg)
			route, err := h.Registrate()
			assert.NoError(t, err)
			r := gin.Default()
			r.POST(route.Path, route.Handler)

			srv := httptest.NewServer(r)
			defer srv.Close()

			req, err := http.NewRequest(http.MethodPost, srv.URL+tt.fields.route, nil)
			assert.NoError(t, err)

			resp, err := srv.Client().Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.NoError(t, resp.Body.Close())
		})
	}
}
