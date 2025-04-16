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
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
)

func TestNewRootHandler(t *testing.T) {
	type args struct {
		counter RootCounterRepository
		gauge   RootGaugeRepository
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "creates instance of RootHandler",
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewRootHandler(tt.args.counter, tt.args.gauge)
			assert.NotNil(t, h)
			assert.Equal(t, tt.args.counter, h.counter)
			assert.Equal(t, tt.args.gauge, h.gauge)
		})
	}
}

func TestRootHandler_Registrate(t *testing.T) {
	type args struct {
		counter RootCounterRepository
		gauge   RootGaugeRepository
	}

	tests := []struct {
		name string
		want server.Route
		args args
	}{
		{
			want: server.Route{
				Path:   "/",
				Method: "GET",
			},
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewRootHandler(tt.args.counter, tt.args.gauge)
			route, err := h.Registrate()
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Path, route.Path)
			assert.Equal(t, tt.want.Method, route.Method)
		})
	}
}

func TestRootHandler_handler(t *testing.T) {
	type fields struct {
		counter *repositories.MockICounterRepository
		gauge   *repositories.MockIGaugeRepository
	}
	tests := []struct {
		name           string
		fields         fields
		wantStatusCode int
		prepare        func(f *fields)
	}{
		{
			name:           "when repositoirs retriews data succeded then return status ok",
			fields:         fields{},
			wantStatusCode: http.StatusOK,
			prepare: func(f *fields) {
				f.gauge.EXPECT().All(gomock.Any()).Return([]models.Gauge{}, nil)
				f.counter.EXPECT().All(gomock.Any()).Return([]models.Counter{}, nil)
			},
		},
		{
			name:           "when gauge repositoirs retriews data failed then return status ok",
			fields:         fields{},
			wantStatusCode: http.StatusInternalServerError,
			prepare: func(f *fields) {
				f.gauge.EXPECT().All(gomock.Any()).Return([]models.Gauge{}, errors.New(""))
			},
		},
		{
			name:           "when gauge repositoirs retriews data failed then return status ok",
			fields:         fields{},
			wantStatusCode: http.StatusInternalServerError,
			prepare: func(f *fields) {
				f.gauge.EXPECT().All(gomock.Any()).Return([]models.Gauge{}, nil)
				f.counter.EXPECT().All(gomock.Any()).Return([]models.Counter{}, errors.New(""))
			},
		},
	}

	cntr := gomock.NewController(t)
	defer cntr.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.counter = repositories.NewMockICounterRepository(cntr)
			tt.fields.gauge = repositories.NewMockIGaugeRepository(cntr)

			tt.prepare(&tt.fields)
			h := NewRootHandler(tt.fields.counter, tt.fields.gauge)

			route, err := h.Registrate()
			assert.NoError(t, err)

			r := gin.Default()
			r.SetHTMLTemplate(route.HTMLTemplates[0])
			r.GET(route.Path, h.handler())

			srv := httptest.NewServer(r)
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
			assert.NoError(t, err)

			resp, err := srv.Client().Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

		})
	}
}
