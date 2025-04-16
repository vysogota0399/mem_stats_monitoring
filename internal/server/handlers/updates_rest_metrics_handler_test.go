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
	mock "github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestNewUpdatesRestMetricsHandler(t *testing.T) {
	type args struct {
		service IUpdateMetricsService
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "creates instance of UpdatesRestMetricsHandler",
			args: args{},
		},
	}
	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.service = mock.NewMockIUpdateMetricsService(ctrl)
			h := NewUpdatesRestMetricsHandler(tt.args.service, lg)
			assert.NotNil(t, h)
			assert.Equal(t, tt.args.service, h.service)
			assert.Equal(t, lg, h.lg)
		})
	}
}

func TestUpdatesRestMetricsHandler_Registrate(t *testing.T) {
	type fields struct {
		service IUpdateMetricsService
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
				service: mock.NewMockIUpdateMetricsService(nil),
				lg:      nil,
			},
			want: server.Route{
				Path:   "/updates/",
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
				Path:   "/updates/",
				Method: "POST",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &UpdatesRestMetricsHandler{
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

func TestUpdatesRestMetricsHandler_handler(t *testing.T) {
	type fields struct {
		srv  *mock.MockIUpdateMetricsService
		body io.Reader
	}
	tests := []struct {
		name           string
		fields         fields
		prepare        func(*fields)
		wantStatusCode int
	}{
		{
			name:           "when invalid json",
			fields:         fields{body: strings.NewReader(`[{"id":}]`)},
			wantStatusCode: http.StatusBadRequest,
			prepare:        func(fields *fields) {},
		},
		{
			name:           "when empty metrics array",
			fields:         fields{body: strings.NewReader(`[]`)},
			wantStatusCode: http.StatusBadRequest,
			prepare:        func(fields *fields) {},
		},
		{
			name: "when successful batch update",
			fields: fields{body: strings.NewReader(`[
				{"id": "test1", "type": "gauge", "value": 1.5},
				{"id": "test2", "type": "counter", "delta": 1}
			]`)},
			prepare: func(fields *fields) {
				fields.srv.EXPECT().Call(gomock.Any(), gomock.Any()).Return(service.UpdateMetricsServiceResult{}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "when service error",
			fields: fields{body: strings.NewReader(`[
				{"id": "test1", "type": "gauge", "value": 1.5}
			]`)},
			prepare: func(fields *fields) {
				fields.srv.EXPECT().Call(gomock.Any(), gomock.Any()).Return(service.UpdateMetricsServiceResult{}, errors.New("service error"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.srv = mock.NewMockIUpdateMetricsService(ctrl)
			tt.prepare(&tt.fields)

			h := NewUpdatesRestMetricsHandler(tt.fields.srv, lg)

			r := gin.Default()
			route, err := h.Registrate()
			assert.NoError(t, err)
			r.POST(route.Path, route.Handler)

			srv := httptest.NewServer(r)
			defer srv.Close()

			req, err := http.NewRequest(http.MethodPost, srv.URL+route.Path, tt.fields.body)
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := srv.Client().Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}
