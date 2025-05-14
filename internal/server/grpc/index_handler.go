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

	"google.golang.org/protobuf/types/known/emptypb"
)

type IMetricsGaugeRepository interface {
	All(ctx context.Context) ([]models.Gauge, error)
}

type IMetricsCounterRepository interface {
	All(ctx context.Context) ([]models.Counter, error)
}

var _ IMetricsGaugeRepository = (*repositories.GaugeRepository)(nil)
var _ IMetricsCounterRepository = (*repositories.CounterRepository)(nil)

type IndexHandler struct {
	gaugeRepository   IMetricsGaugeRepository
	counterRepository IMetricsCounterRepository
	lg                *logging.ZapLogger
}

func NewIndexHandler(gaugeRepository IMetricsGaugeRepository, counterRepository IMetricsCounterRepository, lg *logging.ZapLogger) *IndexHandler {
	return &IndexHandler{gaugeRepository: gaugeRepository, counterRepository: counterRepository, lg: lg}
}

func (h *IndexHandler) Index(ctx context.Context, params *emptypb.Empty) (*metrics.IndexResponse, error) {
	gauges, err := h.gaugeRepository.All(ctx)
	if err != nil {
		h.lg.ErrorCtx(ctx, "fetch gauges failed error", zap.Error(err))
		return nil, status.Error(codes.Internal, "fetch gauges failed error")
	}

	counters, err := h.counterRepository.All(ctx)
	if err != nil {
		h.lg.ErrorCtx(ctx, "fetch counter failed error", zap.Error(err))
		return nil, status.Error(codes.Internal, "fetch counter failed error")
	}

	items := make([]*metrics.Item, 0, len(gauges)+len(counters))
	for _, gg := range gauges {
		item := &metrics.Item_Gauge{
			Gauge: &entities.Gauge{
				Value: gg.Value,
				Name:  gg.Name,
			},
		}

		items = append(items, &metrics.Item{
			Metric: item,
		})
	}

	for _, cc := range counters {
		item := &metrics.Item_Counter{
			Counter: &entities.Counter{
				Value: cc.Value,
				Name:  cc.Name,
			},
		}

		items = append(items, &metrics.Item{
			Metric: item,
		})
	}

	return &metrics.IndexResponse{Items: items}, nil
}
