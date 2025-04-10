package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestNew(t *testing.T) {
	type args struct {
		strg storage.Storage
		lg   *logging.ZapLogger
	}
	tests := []struct {
		name string
		args args
		want *Service
	}{
		{
			name: "when valid storage and logger",
			args: args{
				strg: storage.NewMemory(),
				lg:   func() *logging.ZapLogger { lg, _ := logging.MustZapLogger(-1); return lg }(),
			},
			want: &Service{
				UpdateMetricService: &UpdateMetricService{
					gaugeRep:   nil, // Will be set in the test
					counterRep: nil, // Will be set in the test
				},
				UpdateMetricsService: &UpdateMetricsService{
					gaugeRep:   nil, // Will be set in the test
					counterRep: nil, // Will be set in the test
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.strg, tt.args.lg)
			assert.NotNil(t, got)
			assert.NotNil(t, got.UpdateMetricService)
			assert.NotNil(t, got.UpdateMetricsService)
			assert.NotNil(t, got.UpdateMetricService.gaugeRep)
			assert.NotNil(t, got.UpdateMetricService.counterRep)
			assert.NotNil(t, got.UpdateMetricsService.gaugeRep)
			assert.NotNil(t, got.UpdateMetricsService.counterRep)
		})
	}
}
