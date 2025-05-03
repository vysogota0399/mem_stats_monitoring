package grpc

import (
	"context"
	"testing"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/services/metrics"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func RunTestServer(t testing.TB, config *config.Config, lg *logging.ZapLogger, h metrics.MetricsServiceServer) {
	t.Helper()
	var (
		l fx.Lifecycle
		s fx.Shutdowner
	)
	app := fxtest.New(
		t,
		fx.Populate(&l, &s),
	)

	NewServer(l, config, lg, h)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		t.Errorf("error starting grpc server %s", err.Error())
	}

	t.Cleanup(func() {
		if err := app.Stop(ctx); err != nil {
			t.Errorf("error stopping grpc server %s", err.Error())
		}
	})
}
