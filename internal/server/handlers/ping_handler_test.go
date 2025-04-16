package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock "github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestNewPingHandler(t *testing.T) {
	type args struct {
		strg *mock.MockStorage
	}
	tests := []struct {
		name           string
		args           args
		wantStatusCode int
		prepare        func(*args)
	}{
		{
			name: "when ping failed",
			args: args{},
			prepare: func(args *args) {
				args.strg.EXPECT().Ping(gomock.Any()).Return(errors.New("ping failed"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "when ping ok",
			args: args{},
			prepare: func(args *args) {
				args.strg.EXPECT().Ping(gomock.Any()).Return(nil)
			},
			wantStatusCode: http.StatusOK,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			tt.args.strg = mock.NewMockStorage(ctrl)
			tt.prepare(&tt.args)
			h := NewPingHandler(tt.args.strg, lg)
			r.GET("/ping", h.handler())

			srv := httptest.NewServer(r)
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL+"/ping", nil)
			assert.NoError(t, err)

			resp, err := srv.Client().Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}

func TestPingHandler_Registrate(t *testing.T) {
	type fields struct {
		strg storages.Storage
	}
	tests := []struct {
		name   string
		fields fields
		want   server.Route
	}{
		{
			name: "when storage is not pg",
			fields: fields{
				strg: storages.NewMemory(nil),
			},
			want: server.Route{},
		},
		{
			name: "when storage is pg",
			fields: fields{
				strg: &storages.PG{},
			},
			want: server.Route{
				Path:   "/ping",
				Method: "GET",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewPingHandler(tt.fields.strg, nil)
			route, err := h.Registrate()
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Path, route.Path)
			assert.Equal(t, tt.want.Method, route.Method)
		})
	}
}
