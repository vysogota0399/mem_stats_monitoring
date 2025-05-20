package grpc

import (
	"context"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/entities"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/services/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IShowMetricGaugeRepository interface {
	FindByName(ctx context.Context, name string) (models.Gauge, error)
}

type IShowMetricCounterRepository interface {
	FindByName(ctx context.Context, name string) (models.Counter, error)
}

var _ IShowMetricGaugeRepository = (*repositories.GaugeRepository)(nil)
var _ IShowMetricCounterRepository = (*repositories.CounterRepository)(nil)

type ShowHandler struct {
	gaugeRepository   IShowMetricGaugeRepository
	counterRepository IShowMetricCounterRepository
	lg                *logging.ZapLogger
}

func NewShowHandler(gaugeRepository IShowMetricGaugeRepository, counterRepository IShowMetricCounterRepository, lg *logging.ZapLogger) *ShowHandler {
	return &ShowHandler{gaugeRepository: gaugeRepository, counterRepository: counterRepository, lg: lg}
}

func (h *ShowHandler) Show(ctx context.Context, params *metrics.ShowMetricParams) (*metrics.ShowMetricResponse, error) {
	ctx = h.lg.WithContextFields(ctx, zap.String("handler", "show_handler"))
	switch params.MType {
	case entities.MetricTypes_COUNTER:
		m, err := h.counterRepository.FindByName(ctx, params.Name)
		if err != nil {
			h.lg.ErrorCtx(ctx, "find counter error", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return &metrics.ShowMetricResponse{
			Items: &metrics.Item{
				Metric: &metrics.Item_Counter{
					Counter: &entities.Counter{
						Value: m.Value,
						Name:  m.Name,
					},
				},
			},
		}, nil
	case entities.MetricTypes_GAUGE:
		m, err := h.gaugeRepository.FindByName(ctx, params.Name)
		if err != nil {
			h.lg.ErrorCtx(ctx, "find gauge error", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return &metrics.ShowMetricResponse{
			Items: &metrics.Item{
				Metric: &metrics.Item_Gauge{
					Gauge: &entities.Gauge{
						Value: m.Value,
						Name:  m.Name,
					},
				},
			},
		}, nil
	}

	return nil, nil
}
