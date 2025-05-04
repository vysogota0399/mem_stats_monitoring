package grpc

import (
	"context"
	"testing"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/services/metrics"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

type TestHandler struct {
	*Handler
}

type TestHandlerOpt func(h *TestHandler)

func NewTestHandler(t testing.TB, opts ...TestHandlerOpt) *TestHandler {
	t.Helper()
	base := NewHandler()
	h := &TestHandler{Handler: base}
	for _, f := range opts {
		f(h)
	}

	return h
}

func NewTestClient(t testing.TB, cfg *config.Config) metrics.MetricsServiceClient {
	t.Helper()

	conn, err := grpc.NewClient(":"+cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Errorf("failed to open connection, maybe server is not running")
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Errorf("failed to close connection")
		}
	})

	return metrics.NewMetricsServiceClient(conn)
}
