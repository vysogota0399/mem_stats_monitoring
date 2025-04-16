package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestNewShowMetricHandler(t *testing.T) {
	type args struct {
		gaugeRepository   *repositories.MockIGaugeRepository
		counterRepository *repositories.MockICounterRepository
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "creates instance of ShowMetricHandler",
			args: args{},
		},
	}
	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.gaugeRepository = repositories.NewMockIGaugeRepository(ctrl)
			tt.args.counterRepository = repositories.NewMockICounterRepository(ctrl)

			h := NewShowMetricHandler(tt.args.gaugeRepository, tt.args.counterRepository, lg)
			assert.NotNil(t, h)
			assert.Equal(t, tt.args.gaugeRepository, h.gaugeRepository)
			assert.Equal(t, tt.args.counterRepository, h.counterRepository)
			assert.Equal(t, lg, h.lg)
		})
	}
}

func TestShowMetricHandler_Registrate(t *testing.T) {
	type fields struct {
		gaugeRepository   IShowMetricGaugeRepository
		counterRepository IShowMetricCounterRepository
		lg                *logging.ZapLogger
	}
	tests := []struct {
		name   string
		fields fields
		want   server.Route
	}{
		{
			name: "when repositories are provided",
			fields: fields{
				gaugeRepository:   &repositories.MockIGaugeRepository{},
				counterRepository: &repositories.MockICounterRepository{},
				lg:                nil,
			},
			want: server.Route{
				Path:   "/value/:type/:name",
				Method: "GET",
			},
		},
		{
			name: "when repositories are nil",
			fields: fields{
				gaugeRepository:   nil,
				counterRepository: nil,
				lg:                nil,
			},
			want: server.Route{
				Path:   "/value/:type/:name",
				Method: "GET",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ShowMetricHandler{
				gaugeRepository:   tt.fields.gaugeRepository,
				counterRepository: tt.fields.counterRepository,
				lg:                tt.fields.lg,
			}
			got, err := h.Registrate()
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Path, got.Path)
			assert.Equal(t, tt.want.Method, got.Method)
			assert.NotNil(t, got.Handler)
		})
	}
}

func TestShowMetricHandler_handler(t *testing.T) {
	type fields struct {
		gaugeRepository   *repositories.MockIGaugeRepository
		counterRepository *repositories.MockICounterRepository
		route             string
	}
	tests := []struct {
		name           string
		fields         fields
		want           gin.HandlerFunc
		prepare        func(*fields)
		wantStatusCode int
	}{
		{
			name:           "when invalid metric type",
			fields:         fields{route: "/value/invalid/test"},
			wantStatusCode: http.StatusNotFound,
			prepare:        func(fields *fields) {},
		},
		{
			name:   "when counter metric found",
			fields: fields{route: "/value/counter/test"},
			prepare: func(fields *fields) {
				fields.counterRepository.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(models.Counter{
					Name:  "test",
					Value: 1,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "when gauge metric found",
			fields: fields{route: "/value/gauge/test"},
			prepare: func(fields *fields) {
				fields.gaugeRepository.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(models.Gauge{
					Name:  "test",
					Value: 1,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "when counter metric error occurs",
			fields:         fields{route: "/value/counter/test"},
			wantStatusCode: http.StatusInternalServerError,
			prepare: func(fields *fields) {
				fields.counterRepository.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(models.Counter{}, errors.New("metric not found"))
			},
		},
		{
			name:           "when gauge metric error occurs",
			fields:         fields{route: "/value/gauge/test"},
			wantStatusCode: http.StatusInternalServerError,
			prepare: func(fields *fields) {
				fields.gaugeRepository.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(models.Gauge{}, errors.New("metric not found"))
			},
		},
		{
			name:           "when counter metric not found",
			fields:         fields{route: "/value/counter/test"},
			wantStatusCode: http.StatusNotFound,
			prepare: func(fields *fields) {
				fields.counterRepository.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(models.Counter{}, storages.ErrNoRecords)
			},
		},
		{
			name:           "when gauge metric not found",
			fields:         fields{route: "/value/gauge/test"},
			wantStatusCode: http.StatusNotFound,
			prepare: func(fields *fields) {
				fields.gaugeRepository.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(models.Gauge{}, storages.ErrNoRecords)
			},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.counterRepository = repositories.NewMockICounterRepository(ctrl)
			tt.fields.gaugeRepository = repositories.NewMockIGaugeRepository(ctrl)
			tt.prepare(&tt.fields)

			h := &ShowMetricHandler{
				gaugeRepository:   tt.fields.gaugeRepository,
				counterRepository: tt.fields.counterRepository,
				lg:                lg,
			}

			r := gin.Default()
			r.GET("/value/:type/:name", h.handler())

			srv := httptest.NewServer(r)
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL+tt.fields.route, nil)
			assert.NoError(t, err)

			resp, err := srv.Client().Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}
