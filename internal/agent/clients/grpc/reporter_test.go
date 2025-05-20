package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/services/metrics/mocks"
)

func TestNewReporter(t *testing.T) {
	type args struct {
		ctx context.Context
		cfg *config.Config
		rep *agent.MetricsRepository
	}
	tests := []struct {
		name    string
		args    args
		want    *Reporter
		wantErr bool
	}{
		{
			name: "successful reporter creation",
			args: args{
				ctx: context.Background(),
				cfg: &config.Config{
					GRPCPort: "3200",
				},
				rep: agent.NewMetricsRepository(storage.NewMemoryStorage(nil)),
			},
			wantErr: false,
		},
	}

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			got, err := NewReporter(ctx, tt.args.cfg, tt.args.rep, lg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewReporter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			cancel()
			if !tt.wantErr {
				assert.NotNil(t, got)
				assert.NotNil(t, got.client)
				assert.Equal(t, lg, got.lg)
				assert.Equal(t, tt.args.rep, got.rep)
			}
		})
	}
}

func TestReporter_UpdateMetric(t *testing.T) {
	type fields struct {
		client *mocks.MockMetricsServiceClient
	}
	type args struct {
		ctx   context.Context
		mType string
		mName string
		value string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(f *fields)
		wantErr bool
	}{
		{
			name:   "successful update counter",
			fields: fields{},
			args: args{
				ctx:   context.Background(),
				mType: models.CounterType,
				mName: "test",
				value: "1",
			},
			prepare: func(f *fields) {
				f.client.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name:   "successful update gauge",
			fields: fields{},
			args: args{
				ctx:   context.Background(),
				mType: models.GaugeType,
				mName: "test",
				value: "1",
			},
			prepare: func(f *fields) {
				f.client.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name:   "failed update metric",
			fields: fields{},
			args: args{
				ctx:   context.Background(),
				mType: models.CounterType,
				mName: "test",
				value: "1",
			},
			prepare: func(f *fields) {
				f.client.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
	}

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	cntr := gomock.NewController(t)
	defer cntr.Finish()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.client = mocks.NewMockMetricsServiceClient(cntr)
			tt.prepare(&tt.fields)
			r := &Reporter{
				client: tt.fields.client,
				lg:     lg,
				rep:    agent.NewMetricsRepository(storage.NewMemoryStorage(nil)),
			}

			err := r.UpdateMetric(tt.args.ctx, tt.args.mType, tt.args.mName, tt.args.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReporter_UpdateMetrics(t *testing.T) {
	type fields struct {
		client *mocks.MockMetricsServiceClient
	}
	type args struct {
		data []*models.Metric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(f *fields)
		wantErr bool
	}{
		{
			name:   "successful update metrics",
			fields: fields{},
			args: args{
				data: []*models.Metric{
					{
						Name:  "test",
						Type:  models.CounterType,
						Value: "1",
					},
					{
						Name:  "test2",
						Type:  models.GaugeType,
						Value: "2",
					},
				},
			},
			prepare: func(f *fields) {
				f.client.EXPECT().UpdateBatch(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name:   "failed update metrics",
			fields: fields{},
			args: args{
				data: []*models.Metric{
					{
						Name:  "test",
						Type:  models.CounterType,
						Value: "1",
					},
				},
			},
			prepare: func(f *fields) {
				f.client.EXPECT().UpdateBatch(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
	}

	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	cntr := gomock.NewController(t)
	defer cntr.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.client = mocks.NewMockMetricsServiceClient(cntr)
			tt.prepare(&tt.fields)
			r := &Reporter{
				client: tt.fields.client,
				lg:     lg,
				rep:    agent.NewMetricsRepository(storage.NewMemoryStorage(nil)),
			}
			err := r.UpdateMetrics(context.Background(), tt.args.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
