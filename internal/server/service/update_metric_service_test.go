package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

func TestUpdateMetricService_Call(t *testing.T) {
	var value = 1.0
	var delta int64 = 1
	s := storage.NewMemory()
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
				MType: "counter",
				Value: &value,
				Delta: &delta,
			},
			wantErr: false,
			want: UpdateMetricServiceResult{
				ID:    "test",
				MType: "counter",
				Delta: &delta,
			},
		},
		{
			name: "when counter without delta then returns error",
			args: UpdateMetricServiceParams{
				MName: "test",
				MType: "counter",
				Delta: nil,
			},
			wantErr: true,
			want:    UpdateMetricServiceResult{},
		},
		{
			name: "when gauge with value returns new value",
			args: UpdateMetricServiceParams{
				MName: "test",
				MType: "gauge",
				Value: &value,
			},
			wantErr: false,
			want: UpdateMetricServiceResult{
				ID:    "test",
				MType: "gauge",
				Value: &value,
			},
		},
		{
			name: "when gauge without value then returns error",
			args: UpdateMetricServiceParams{
				MName: "test",
				MType: "counter",
				Value: nil,
			},
			wantErr: true,
			want:    UpdateMetricServiceResult{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counterRep := repositories.NewCounter(s)
			gaugeRep := repositories.NewGauge(s)
			s := UpdateMetricService{
				counterRep: counterRep,
				gaugeRep:   gaugeRep,
			}
			res, err := s.Call(tt.args)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, res)
		})
	}
}
