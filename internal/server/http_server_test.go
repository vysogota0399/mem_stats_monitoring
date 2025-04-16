// Package server handles the initialization and operation of the web server.
// It defines endpoints, handlers, and middleware for the metrics collection service.
package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestNewHTTPServer(t *testing.T) {
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
				cfg: &config.Config{
					Address: ":8080",
				},
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

			got := NewHTTPServer(l, tt.args.cfg, NewRouter([]Handler{&MockHandler{}}, l, lg, tt.args.cfg, nil), lg)
			assert.NotNil(t, got)

			err := app.Start(context.Background())
			assert.NoError(t, err)

			err = app.Stop(context.Background())
			assert.NoError(t, err)
		})
	}
}
