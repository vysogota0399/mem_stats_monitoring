package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

func TestUpdateMetricService_Call(t *testing.T) {
	var value = 1.0
	var delta int64 = 1
	var zeroValue float64
	var zeroDelta int64
	tests := []struct {
		name    string
		args    UpdateMetricServiceParams
		want    UpdateMetricServiceResult
		wantErr bool
	}{
		{
			name: "when counter with delta returns new value",
			args: UpdateMetricServiceParams{
				MName: "test",
				MType: models.CounterType,
				Value: &value,
				Delta: &delta,
			},
			wantErr: false,
			want: UpdateMetricServiceResult{
				ID:    "test",
				MType: models.CounterType,
				Delta: &delta,
			},
		},
		{
			name: "when counter without delta then returns no error",
			args: UpdateMetricServiceParams{
				MName: "test",
				MType: models.CounterType,
			},
			wantErr: false,
			want: UpdateMetricServiceResult{
				ID:    "test",
				MType: models.CounterType,
				Delta: &zeroDelta,
			},
		},
		{
			name: "when gauge with value returns new value",
			args: UpdateMetricServiceParams{
				MName: "test",
				MType: models.GaugeType,
				Value: &value,
			},
			wantErr: false,
			want: UpdateMetricServiceResult{
				ID:    "test",
				MType: models.GaugeType,
				Value: &value,
			},
		},
		{
			name: "when gauge without value then returns no error",
			args: UpdateMetricServiceParams{
				MName: "test",
				MType: models.GaugeType,
			},
			wantErr: false,
			want: UpdateMetricServiceResult{
				ID:    "test",
				MType: models.GaugeType,
				Value: &zeroValue,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := storage.NewMemory()

			counterRep := repositories.NewCounter(s)
			gaugeRep := repositories.NewGauge(s)
			serv := UpdateMetricService{
				counterRep: counterRep,
				gaugeRep:   gaugeRep,
			}
			res, err := serv.Call(context.Background(), tt.args)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, res)
		})
	}
}
