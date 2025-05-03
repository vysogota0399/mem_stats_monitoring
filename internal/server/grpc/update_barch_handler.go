package grpc

import (
	"context"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/services/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type IUpdateMetricsService interface {
	Call(context.Context, service.UpdateMetricsServiceParams) (service.UpdateMetricsServiceResult, error)
}

var _ IUpdateMetricsService = (*service.UpdateMetricsService)(nil)

type UpdateBatchHandler struct {
	lg      *logging.ZapLogger
	service IUpdateMetricsService
}

func NewUpdateBatchHandler(lg *logging.ZapLogger, service IUpdateMetricsService) *UpdateBatchHandler {
	return &UpdateBatchHandler{lg: lg, service: service}
}

func (h *UpdateBatchHandler) UpdateBatch(ctx context.Context, params *metrics.UpdateMetricsBatchParams) (*emptypb.Empty, error) {
	ctx = h.lg.WithContextFields(ctx, zap.String("handler", "update_batch_handler"))

	_, err := h.service.Call(ctx, h.prepareParams(params))
	if err != nil {
		h.lg.ErrorCtx(ctx, "update metrics batch error", zap.Error(err))
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *UpdateBatchHandler) prepareParams(params *metrics.UpdateMetricsBatchParams) service.UpdateMetricsServiceParams {
	elements := make([]service.UpdateMetricsServiceEl, len(params.Item))

	for i, item := range params.Item {
		switch el := item.Metric.(type) {
		case *metrics.Item_Counter:
			elements[i] = service.UpdateMetricsServiceEl{
				ID:    el.Counter.Name,
				MType: models.CounterType,
				Delta: el.Counter.Value,
			}
		case *metrics.Item_Gauge:
			elements[i] = service.UpdateMetricsServiceEl{
				ID:    el.Gauge.Name,
				MType: models.GaugeType,
				Value: el.Gauge.Value,
			}
		}
	}

	return elements
}
