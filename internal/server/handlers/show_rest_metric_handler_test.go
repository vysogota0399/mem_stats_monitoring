package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestNewShowRestMetricHandler(t *testing.T) {
	type args struct {
		gaugeRepository   *repositories.MockIGaugeRepository
		counterRepository *repositories.MockICounterRepository
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "creates instance of ShowRestMetricHandler",
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

			h := NewShowRestMetricHandler(tt.args.gaugeRepository, tt.args.counterRepository, lg)
			assert.NotNil(t, h)
			assert.Equal(t, tt.args.gaugeRepository, h.gaugeRepository)
			assert.Equal(t, tt.args.counterRepository, h.counterRepository)
			assert.Equal(t, lg, h.lg)
		})
	}
}

func TestShowRestMetricHandler_Registrate(t *testing.T) {
	type fields struct {
		gaugeRepository   IShowRestMetricGaugeRepository
		counterRepository IShowRestMetricCounterRepository
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
				Path:   "/value/",
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
				Path:   "/value/",
				Method: "GET",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ShowRestMetricHandler{
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

func TestShowRestMetricHandler_handler(t *testing.T) {
	type fields struct {
		gaugeRepository   *repositories.MockIGaugeRepository
		counterRepository *repositories.MockICounterRepository
		body              io.Reader
	}
	type want struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(*fields)
		want    want
	}{
		{
			name: "when invalid json",
			fields: fields{
				body: strings.NewReader(`{"id":}`),
			},
			prepare: func(fields *fields) {},
			want: want{
				statusCode: http.StatusBadRequest,
				body:       `{}`,
			},
		},
		{
			name: "when metric not found",
			fields: fields{
				body: strings.NewReader(`{"id": "test", "type": "gauge"}`),
			},
			prepare: func(fields *fields) {
				fields.gaugeRepository.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(models.Gauge{}, storages.ErrNoRecords)
			},
			want: want{
				statusCode: http.StatusNotFound,
				body:       `{}`,
			},
		},
		{
			name: "when error occurs",
			fields: fields{
				body: strings.NewReader(`{"id": "test", "type": "gauge"}`),
			},
			prepare: func(fields *fields) {
				fields.gaugeRepository.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(models.Gauge{}, errors.New("error"))
			},
			want: want{
				statusCode: http.StatusBadRequest,
				body:       `{}`,
			},
		},
		{
			name: "when gauge metric found",
			fields: fields{
				body: strings.NewReader(`{"id": "test", "type": "gauge"}`),
			},
			prepare: func(fields *fields) {
				fields.gaugeRepository.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(models.Gauge{
					Name:  "test",
					Value: 1,
				}, nil)
			},
			want: want{
				statusCode: http.StatusOK,
				body:       `{"id": "test", "type": "gauge", "value": 1}`,
			},
		},
		{
			name: "when counter metric found",
			fields: fields{
				body: strings.NewReader(`{"id": "test", "type": "counter"}`),
			},
			prepare: func(fields *fields) {
				fields.counterRepository.EXPECT().FindByName(gomock.Any(), gomock.Any()).Return(models.Counter{
					Name:  "test",
					Value: 1,
				}, nil)
			},
			want: want{
				statusCode: http.StatusOK,
				body:       `{"id": "test", "type": "counter", "delta": 1}`,
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.gaugeRepository = repositories.NewMockIGaugeRepository(ctrl)
			tt.fields.counterRepository = repositories.NewMockICounterRepository(ctrl)

			tt.prepare(&tt.fields)

			h := NewShowRestMetricHandler(tt.fields.gaugeRepository, tt.fields.counterRepository, lg)

			r := gin.Default()
			route, err := h.Registrate()
			assert.NoError(t, err)
			r.POST(route.Path, route.Handler)

			srv := httptest.NewServer(r)
			defer srv.Close()

			req, err := http.NewRequest(http.MethodPost, srv.URL+route.Path, tt.fields.body)
			assert.NoError(t, err)

			resp, err := srv.Client().Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.want.body, string(body))
		})
	}
}
