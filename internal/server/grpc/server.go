package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"google.golang.org/grpc"

	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/services/metrics"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Server struct {
	cfg        *config.Config
	lg         *logging.ZapLogger
	grpcServer *grpc.Server
}

var _ metrics.MetricsServiceServer = (*Handler)(nil)

func NewServer(
	lc fx.Lifecycle,
	cfg *config.Config,
	lg *logging.ZapLogger,
	handler metrics.MetricsServiceServer,
) *Server {
	srv := &Server{
		cfg: cfg,
		lg:  lg,
	}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				ln, err := net.Listen("tcp", ":"+cfg.GRPCPort)
				if err != nil {
					return fmt.Errorf("server: failed to create listener on port %s error %w", cfg.GRPCPort, err)
				}

				srv.grpcServer = grpc.NewServer()

				metrics.RegisterMetricsServiceServer(srv.grpcServer, handler)
				go func() {
					if err := srv.grpcServer.Serve(ln); err != nil {
						srv.lg.ErrorCtx(ctx, "server: serve failer error", zap.Error(err))
					}
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {
				srv.grpcServer.GracefulStop()
				return nil
			},
		},
	)

	return srv
}
