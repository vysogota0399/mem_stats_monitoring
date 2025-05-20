package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock "github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/entities"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/services/metrics"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestNewUpdateBatchHandler(t *testing.T) {
	type args struct {
		service IUpdateMetricsService
	}
	tests := []struct {
		name string
		args args
		want *UpdateBatchHandler
	}{
		{
			name: "creates instance of UpdateBatchHandler",
			args: args{
				service: nil,
			},
			want: &UpdateBatchHandler{
				lg:      nil,
				service: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewUpdateBatchHandler(nil, tt.args.service)
			assert.Equal(t, tt.want.service, got.service)
			assert.Equal(t, tt.want.lg, got.lg)
		})
	}
}

func TestUpdateBatchHandler_UpdateBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	cfg := &config.Config{
		GRPCPort: "3200",
	}

	type fields struct {
		service *mock.MockIUpdateMetricsService
	}
	type args struct {
		params *metrics.UpdateMetricsBatchParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(*fields)
		want    *emptypb.Empty
		wantErr bool
	}{
		{
			name:   "when service error",
			fields: fields{},
			args: args{
				params: &metrics.UpdateMetricsBatchParams{
					Item: []*metrics.Item{
						{
							Metric: &metrics.Item_Gauge{
								Gauge: &entities.Gauge{
									Name:  "test1",
									Value: 1.5,
								},
							},
						},
					},
				},
			},
			wantErr: true,
			prepare: func(f *fields) {
				f.service.EXPECT().Call(gomock.Any(), gomock.Any()).Return(service.UpdateMetricsServiceResult{}, errors.New("service error"))
			},
		},
		{
			name:   "when successful batch update",
			fields: fields{},
			args: args{
				params: &metrics.UpdateMetricsBatchParams{
					Item: []*metrics.Item{
						{
							Metric: &metrics.Item_Gauge{
								Gauge: &entities.Gauge{
									Name:  "test1",
									Value: 1.5,
								},
							},
						},
						{
							Metric: &metrics.Item_Counter{
								Counter: &entities.Counter{
									Name:  "test2",
									Value: 1,
								},
							},
						},
					},
				},
			},
			prepare: func(f *fields) {
				f.service.EXPECT().Call(gomock.Any(), gomock.Any()).Return(service.UpdateMetricsServiceResult{}, nil)
			},
			want: &emptypb.Empty{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.service = mock.NewMockIUpdateMetricsService(ctrl)
			tt.prepare(&tt.fields)

			handler := NewUpdateBatchHandler(lg, tt.fields.service)
			th := NewTestHandler(t, func(h *TestHandler) {
				h.UpdateBatchHandler = *handler
			})

			RunTestServer(t, cfg, lg, th)

			client := NewTestClient(t, cfg)
			resp, err := client.UpdateBatch(context.Background(), tt.args.params)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}
