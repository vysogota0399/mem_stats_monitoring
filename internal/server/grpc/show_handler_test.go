package grpc

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/entities"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/services/metrics"
)

func TestNewShowHandler(t *testing.T) {
	type args struct {
		gaugeRepository   IShowMetricGaugeRepository
		counterRepository IShowMetricCounterRepository
		lg                *logging.ZapLogger
	}
	tests := []struct {
		name string
		args args
		want *ShowHandler
	}{
		{
			name: "creates instance of ShowHandler",
			args: args{
				gaugeRepository:   nil,
				counterRepository: nil,
				lg:                nil,
			},
			want: &ShowHandler{
				gaugeRepository:   nil,
				counterRepository: nil,
				lg:                nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewShowHandler(tt.args.gaugeRepository, tt.args.counterRepository, tt.args.lg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewShowHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShowHandler_Show(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	cfg := &config.Config{
		GRPCPort: "3200",
	}

	type fields struct {
		gaugeRepository   *repositories.MockIGaugeRepository
		counterRepository *repositories.MockICounterRepository
	}
	type args struct {
		params *metrics.ShowMetricParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(*fields)
		want    *metrics.ShowMetricResponse
		wantErr bool
	}{
		{
			name:   "when gauge metric found",
			fields: fields{},
			args: args{
				params: &metrics.ShowMetricParams{
					Name:  "test",
					MType: entities.MetricTypes_GAUGE,
				},
			},
			prepare: func(f *fields) {
				f.gaugeRepository.EXPECT().FindByName(gomock.Any(), "test").Return(models.Gauge{
					Name:  "test",
					Value: 1.5,
				}, nil)
			},
			want: &metrics.ShowMetricResponse{
				Items: &metrics.Item{
					Metric: &metrics.Item_Gauge{
						Gauge: &entities.Gauge{
							Name:  "test",
							Value: 1.5,
						},
					},
				},
			},
		},
		{
			name:   "when counter metric found",
			fields: fields{},
			args: args{
				params: &metrics.ShowMetricParams{
					Name:  "test",
					MType: entities.MetricTypes_COUNTER,
				},
			},
			prepare: func(f *fields) {
				f.counterRepository.EXPECT().FindByName(gomock.Any(), "test").Return(models.Counter{
					Name:  "test",
					Value: 1,
				}, nil)
			},
			want: &metrics.ShowMetricResponse{
				Items: &metrics.Item{
					Metric: &metrics.Item_Counter{
						Counter: &entities.Counter{
							Name:  "test",
							Value: 1,
						},
					},
				},
			},
		},
		{
			name:   "when gauge metric not found",
			fields: fields{},
			args: args{
				params: &metrics.ShowMetricParams{
					Name:  "test",
					MType: entities.MetricTypes_GAUGE,
				},
			},
			prepare: func(f *fields) {
				f.gaugeRepository.EXPECT().FindByName(gomock.Any(), "test").Return(models.Gauge{}, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name:   "when counter metric not found",
			fields: fields{},
			args: args{
				params: &metrics.ShowMetricParams{
					Name:  "test",
					MType: entities.MetricTypes_COUNTER,
				},
			},
			prepare: func(f *fields) {
				f.counterRepository.EXPECT().FindByName(gomock.Any(), "test").Return(models.Counter{}, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.gaugeRepository = repositories.NewMockIGaugeRepository(ctrl)
			tt.fields.counterRepository = repositories.NewMockICounterRepository(ctrl)
			tt.prepare(&tt.fields)

			handler := NewShowHandler(tt.fields.gaugeRepository, tt.fields.counterRepository, lg)
			th := NewTestHandler(t, func(h *TestHandler) {
				h.ShowHandler = *handler
			})

			RunTestServer(t, cfg, lg, th)

			client := NewTestClient(t, cfg)
			resp, err := client.Show(context.Background(), tt.args.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				switch tt.args.params.MType {
				case entities.MetricTypes_GAUGE:
					assert.Equal(t, tt.want.Items.GetGauge().Name, resp.Items.GetGauge().Name)
					assert.Equal(t, tt.want.Items.GetGauge().Value, resp.Items.GetGauge().Value)
				case entities.MetricTypes_COUNTER:
					assert.Equal(t, tt.want.Items.GetCounter().Name, resp.Items.GetCounter().Name)
					assert.Equal(t, tt.want.Items.GetCounter().Value, resp.Items.GetCounter().Value)
				}
			}
		})
	}
}
