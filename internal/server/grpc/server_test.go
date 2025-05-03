package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestNewServer(t *testing.T) {
	type args struct {
		cfg *config.Config
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "no errors",
			args: args{
				cfg: &config.Config{},
			},
		},
	}

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	if err != nil {
		t.Fatalf("failed to create zap logger: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				l fx.Lifecycle
				s fx.Shutdowner
			)
			app := fxtest.New(
				t,
				fx.Populate(&l, &s),
			)

			got := NewServer(l, tt.args.cfg, lg, &Handler{})
			assert.NotNil(t, got)

			err := app.Start(context.Background())
			assert.NoError(t, err)

			err = app.Stop(context.Background())
			assert.NoError(t, err)
		})
	}
}
