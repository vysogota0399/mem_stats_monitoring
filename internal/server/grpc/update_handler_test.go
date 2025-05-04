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

func TestNewUpdateHandler(t *testing.T) {
	type args struct {
		service IUpdateMetricService
	}
	tests := []struct {
		name string
		args args
		want *UpdateHandler
	}{
		{
			name: "creates instance of UpdateHandler",
			args: args{
				service: nil,
			},
			want: &UpdateHandler{
				lg:      nil,
				service: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewUpdateHandler(nil, tt.args.service)
			assert.Equal(t, tt.want.service, got.service)
			assert.Equal(t, tt.want.lg, got.lg)
		})
	}
}

func TestUpdateHandler_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	cfg := &config.Config{
		GRPCPort: "3200",
	}

	type fields struct {
		service *mock.MockIUpdateMetricService
	}
	type args struct {
		params *metrics.UpdateMetricParams
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
			name:   "when service error for gauge",
			fields: fields{},
			args: args{
				params: &metrics.UpdateMetricParams{
					Item: &metrics.Item{
						Metric: &metrics.Item_Gauge{
							Gauge: &entities.Gauge{
								Name:  "test1",
								Value: 1.5,
							},
						},
					},
				},
			},
			wantErr: true,
			prepare: func(f *fields) {
				f.service.EXPECT().Call(gomock.Any(), gomock.Any()).Return(service.UpdateMetricServiceResult{}, errors.New("service error"))
			},
		},
		{
			name:   "when service error for counter",
			fields: fields{},
			args: args{
				params: &metrics.UpdateMetricParams{
					Item: &metrics.Item{
						Metric: &metrics.Item_Counter{
							Counter: &entities.Counter{
								Name:  "test2",
								Value: 1,
							},
						},
					},
				},
			},
			wantErr: true,
			prepare: func(f *fields) {
				f.service.EXPECT().Call(gomock.Any(), gomock.Any()).Return(service.UpdateMetricServiceResult{}, errors.New("service error"))
			},
		},
		{
			name:   "when successful gauge update",
			fields: fields{},
			args: args{
				params: &metrics.UpdateMetricParams{
					Item: &metrics.Item{
						Metric: &metrics.Item_Gauge{
							Gauge: &entities.Gauge{
								Name:  "test1",
								Value: 1.5,
							},
						},
					},
				},
			},
			prepare: func(f *fields) {
				f.service.EXPECT().Call(gomock.Any(), gomock.Any()).Return(service.UpdateMetricServiceResult{}, nil)
			},
			want: &emptypb.Empty{},
		},
		{
			name:   "when successful counter update",
			fields: fields{},
			args: args{
				params: &metrics.UpdateMetricParams{
					Item: &metrics.Item{
						Metric: &metrics.Item_Counter{
							Counter: &entities.Counter{
								Name:  "test2",
								Value: 1,
							},
						},
					},
				},
			},
			prepare: func(f *fields) {
				f.service.EXPECT().Call(gomock.Any(), gomock.Any()).Return(service.UpdateMetricServiceResult{}, nil)
			},
			want: &emptypb.Empty{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.service = mock.NewMockIUpdateMetricService(ctrl)
			tt.prepare(&tt.fields)

			handler := NewUpdateHandler(lg, tt.fields.service)
			th := NewTestHandler(t, func(h *TestHandler) {
				h.UpdateHandler = *handler
			})

			RunTestServer(t, cfg, lg, th)

			client := NewTestClient(t, cfg)
			resp, err := client.Update(context.Background(), tt.args.params)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}
