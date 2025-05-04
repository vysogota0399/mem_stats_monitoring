package grpc

import (
	"context"

	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/services/metrics"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	metrics.UnimplementedMetricsServiceServer
	PingHandler
	IndexHandler
	ShowHandler
	UpdateBatchHandler
	UpdateHandler
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Ping(ctx context.Context, params *emptypb.Empty) (*emptypb.Empty, error) {
	return h.PingHandler.Ping(ctx, params)
}

func (h *Handler) Index(ctx context.Context, params *emptypb.Empty) (*metrics.IndexResponse, error) {
	return h.IndexHandler.Index(ctx, params)
}

func (h *Handler) Show(ctx context.Context, params *metrics.ShowMetricParams) (*metrics.ShowMetricResponse, error) {
	return h.ShowHandler.Show(ctx, params)
}

func (h *Handler) UpdateBatch(ctx context.Context, params *metrics.UpdateMetricsBatchParams) (*emptypb.Empty, error) {
	return h.UpdateBatchHandler.UpdateBatch(ctx, params)
}

func (h *Handler) Update(ctx context.Context, params *metrics.UpdateMetricParams) (*emptypb.Empty, error) {
	return h.UpdateHandler.Update(ctx, params)
}
