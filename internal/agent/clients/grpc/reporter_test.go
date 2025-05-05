package grpc

import (
	"context"
	"reflect"
	"testing"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestNewReporter(t *testing.T) {
	type args struct {
		ctx context.Context
		cfg *config.Config
		rep *agent.MetricsRepository
		lg  *logging.ZapLogger
	}
	tests := []struct {
		name    string
		args    args
		want    *Reporter
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewReporter(tt.args.ctx, tt.args.cfg, tt.args.rep, tt.args.lg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewReporter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewReporter() = %v, want %v", got, tt.want)
			}
		})
	}
}
