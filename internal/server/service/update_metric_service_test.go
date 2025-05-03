package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
)

func TestUpdateMetricService_Call(t *testing.T) {
	type args struct {
		params     UpdateMetricServiceParams
		counterRep *repositories.MockICounterRepository
		gaugeRep   *repositories.MockIGaugeRepository
	}
	tests := []struct {
		name    string
		args    args
		prepare func(args *args)
		wantErr bool
		want    UpdateMetricServiceResult
	}{
		{
			name: "when create counter error",
			args: args{
				params: UpdateMetricServiceParams{
					MName: "test",
					MType: models.CounterType,
					Delta: 10,
				},
			},
			prepare: func(args *args) {
				args.counterRep.EXPECT().Create(context.Background(), gomock.Any()).Return(errors.New("create counter error"))
			},
			wantErr: true,
		},
		{
			name: "when create counter success",
			args: args{
				params: UpdateMetricServiceParams{
					MName: "test",
					MType: models.CounterType,
					Delta: 10,
				},
			},
			prepare: func(args *args) {
				args.counterRep.EXPECT().Create(context.Background(), gomock.Any()).DoAndReturn(func(ctx context.Context, cntr *models.Counter) error {
					cntr.Value = 20
					return nil
				})
			},
			wantErr: false,
			want: UpdateMetricServiceResult{
				ID:    "test",
				MType: models.CounterType,
				Delta: 20,
			},
		},
		{
			name: "when create gauge error",
			args: args{
				params: UpdateMetricServiceParams{
					MName: "test",
					MType: models.GaugeType,
					Value: 10,
				},
			},
			prepare: func(args *args) {
				args.gaugeRep.EXPECT().Create(context.Background(), gomock.Any()).Return(errors.New("create gauge error"))
			},
			wantErr: true,
		},
		{
			name: "when create gauge success",
			args: args{
				params: UpdateMetricServiceParams{
					MName: "test",
					MType: models.GaugeType,
					Value: 10,
				},
			},
			prepare: func(args *args) {
				args.gaugeRep.EXPECT().Create(context.Background(), gomock.Any()).DoAndReturn(func(ctx context.Context, gauge *models.Gauge) error {
					gauge.Value = 20
					return nil
				})
			},
			wantErr: false,
			want: UpdateMetricServiceResult{
				ID:    "test",
				MType: models.GaugeType,
				Value: 20,
			},
		},
	}

	cntr := gomock.NewController(t)
	defer cntr.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.counterRep = repositories.NewMockICounterRepository(cntr)
			tt.args.gaugeRep = repositories.NewMockIGaugeRepository(cntr)
			tt.prepare(&tt.args)

			service := NewUpdateMetricService(tt.args.counterRep, tt.args.gaugeRep)
			result, err := service.Call(context.Background(), tt.args.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}
