package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	repoMock "github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestUpdateMetricsService_Call(t *testing.T) {
	type args struct {
		params     UpdateMetricsServiceParams
		counterRep *repoMock.MockICounterRepository
		gaugeRep   *repoMock.MockIGaugeRepository
	}
	tests := []struct {
		name    string
		args    args
		prepare func(args *args)
		wantErr bool
		want    UpdateMetricsServiceResult
	}{
		{
			name: "when save counters error",
			args: args{
				params: UpdateMetricsServiceParams{
					{
						ID:    "test1",
						MType: models.CounterType,
						Delta: 10,
					},
				},
			},
			prepare: func(args *args) {
				args.counterRep.EXPECT().SaveCollection(gomock.Any(), gomock.Any()).Return(errors.New("save counters error"))
			},
			wantErr: true,
		},
		{
			name: "when save gauges error",
			args: args{
				params: UpdateMetricsServiceParams{
					{
						ID:    "test1",
						MType: models.GaugeType,
						Value: 10.5,
					},
				},
			},
			prepare: func(args *args) {
				args.gaugeRep.EXPECT().SaveCollection(gomock.Any(), gomock.Any()).Return(errors.New("save gauges error"))
			},
			wantErr: true,
		},
		{
			name: "when save mixed metrics success",
			args: args{
				params: UpdateMetricsServiceParams{
					{
						ID:    "test1",
						MType: models.CounterType,
						Delta: 10,
					},
					{
						ID:    "test2",
						MType: models.GaugeType,
						Value: 10.5,
					},
				},
			},
			prepare: func(args *args) {
				args.counterRep.EXPECT().SaveCollection(gomock.Any(), gomock.Any()).Return(nil)
				args.gaugeRep.EXPECT().SaveCollection(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
			want:    UpdateMetricsServiceResult{},
		},
	}

	cntr := gomock.NewController(t)
	defer cntr.Finish()

	lg, err := logging.NewZapLogger(&config.Config{})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.counterRep = repoMock.NewMockICounterRepository(cntr)
			tt.args.gaugeRep = repoMock.NewMockIGaugeRepository(cntr)
			tt.prepare(&tt.args)

			rep := NewUpdateMetricsService(
				tt.args.counterRep,
				tt.args.gaugeRep,
				lg,
			)
			got, err := rep.Call(context.Background(), tt.args.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}
