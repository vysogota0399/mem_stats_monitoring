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

type UpdateHandler struct {
	lg      *logging.ZapLogger
	service IUpdateMetricService
}

type IUpdateMetricService interface {
	Call(context.Context, service.UpdateMetricServiceParams) (service.UpdateMetricServiceResult, error)
}

var _ IUpdateMetricService = (*service.UpdateMetricService)(nil)

func NewUpdateHandler(lg *logging.ZapLogger, service IUpdateMetricService) *UpdateHandler {
	return &UpdateHandler{lg: lg, service: service}
}

func (h *UpdateHandler) Update(ctx context.Context, params *metrics.UpdateMetricParams) (*emptypb.Empty, error) {
	ctx = h.lg.WithContextFields(ctx, zap.String("handler", "update_handler"))

	var serviceParams service.UpdateMetricServiceParams

	switch item := params.Item.Metric.(type) {
	case *metrics.Item_Counter:
		serviceParams = service.UpdateMetricServiceParams{
			MName: item.Counter.Name,
			MType: models.CounterType,
			Delta: item.Counter.Value,
		}
	case *metrics.Item_Gauge:
		serviceParams = service.UpdateMetricServiceParams{
			MName: item.Gauge.Name,
			MType: models.GaugeType,
			Value: item.Gauge.Value,
		}
	}

	_, err := h.service.Call(ctx, serviceParams)
	if err != nil {
		h.lg.ErrorCtx(ctx, "error updating metric", zap.Error(err))
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, err.Error())
	}

	return &emptypb.Empty{}, nil
}
