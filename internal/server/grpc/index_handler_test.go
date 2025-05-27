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
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestNewIndexHandler(t *testing.T) {
	type args struct {
		gaugeRepository   IMetricsGaugeRepository
		counterRepository IMetricsCounterRepository
		lg                *logging.ZapLogger
	}
	tests := []struct {
		name string
		args args
		want *IndexHandler
	}{
		{
			name: "creates instance of IndexHandler",
			args: args{
				gaugeRepository:   nil,
				counterRepository: nil,
				lg:                nil,
			},
			want: &IndexHandler{
				gaugeRepository:   nil,
				counterRepository: nil,
				lg:                nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewIndexHandler(tt.args.gaugeRepository, tt.args.counterRepository, tt.args.lg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIndexHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIndexHandler_Index(t *testing.T) {
	type fields struct {
		gaugeRepository   *repositories.MockIGaugeRepository
		counterRepository *repositories.MockICounterRepository
	}
	type args struct {
		params *emptypb.Empty
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(*fields)
		want    *metrics.IndexResponse
		wantErr bool
	}{
		{
			name:    "when fetch gauges failed error",
			fields:  fields{},
			args:    args{params: &emptypb.Empty{}},
			wantErr: true,
			prepare: func(f *fields) {
				f.gaugeRepository.EXPECT().All(gomock.Any()).Return(nil, errors.New("error"))
			},
		},
		{
			name:   "fetch counter failed error",
			fields: fields{},
			args:   args{params: &emptypb.Empty{}},
			prepare: func(f *fields) {
				f.gaugeRepository.EXPECT().All(gomock.Any()).Return([]models.Gauge{}, nil)
				f.counterRepository.EXPECT().All(gomock.Any()).Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
		{
			name:   "when ok",
			fields: fields{},
			args:   args{params: &emptypb.Empty{}},
			prepare: func(f *fields) {
				f.gaugeRepository.EXPECT().All(gomock.Any()).Return([]models.Gauge{{Name: "g1", Value: 1}}, nil)
				f.counterRepository.EXPECT().All(gomock.Any()).Return([]models.Counter{{Name: "c1", Value: 1}}, nil)
			},
			want: &metrics.IndexResponse{
				Items: []*metrics.Item{
					{
						Metric: &metrics.Item_Gauge{
							Gauge: &entities.Gauge{
								Name:  "g1",
								Value: 1,
							},
						},
					},
					{
						Metric: &metrics.Item_Counter{
							Counter: &entities.Counter{
								Name:  "c1",
								Value: 1,
							},
						},
					},
				},
			},
		},
	}

	cfg := &config.Config{
		GRPCPort: "3200",
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

			handler := NewIndexHandler(tt.fields.gaugeRepository, tt.fields.counterRepository, lg)
			th := NewTestHandler(t, func(h *TestHandler) {
				h.IndexHandler = *handler
			})

			RunTestServer(t, cfg, lg, th)

			client := NewTestClient(t, cfg)
			resp, err := client.Index(context.Background(), &emptypb.Empty{})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.want.Items), len(resp.Items))
				for i, item := range resp.Items {
					switch m := item.Metric.(type) {
					case *metrics.Item_Gauge:
						assert.Equal(t, tt.want.Items[i].GetGauge().Name, m.Gauge.Name)
						assert.Equal(t, tt.want.Items[i].GetGauge().Value, m.Gauge.Value)
					case *metrics.Item_Counter:
						assert.Equal(t, tt.want.Items[i].GetCounter().Name, m.Counter.Name)
						assert.Equal(t, tt.want.Items[i].GetCounter().Value, m.Counter.Value)
					}
				}
			}
		})
	}
}
