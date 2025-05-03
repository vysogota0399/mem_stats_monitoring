package grpc

import "github.com/vysogota0399/mem_stats_monitoring/pkg/gen/services/metrics"

type Handler struct {
	metrics.UnimplementedMetricsServiceServer
}
