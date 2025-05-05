package grpc

import (
	"net"
	"testing"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/services/metrics"
	"google.golang.org/grpc"
)

func StartTestServer(t *testing.T, cfg *config.Config) {
	t.Helper()

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	type MetricsServer struct {
		metrics.UnimplementedMetricsServiceServer
	}

	grpcServer := grpc.NewServer()
	metrics.RegisterMetricsServiceServer(grpcServer, &MetricsServer{})

	if err := grpcServer.Serve(lis); err != nil {
		t.Fatalf("failed to serve: %v", err)
	}

	t.Cleanup(func() {
		grpcServer.Stop()
		lis.Close()
	})
}
