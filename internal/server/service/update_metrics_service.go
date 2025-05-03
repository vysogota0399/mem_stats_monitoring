package service

import (
	"context"
	"fmt"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type CntrRep interface {
	SaveCollection(context.Context, []models.Counter) error
}

type GGRep interface {
	SaveCollection(context.Context, []models.Gauge) error
}

// IUpdateMetricsService interface for mocks
type IUpdateMetricsService interface {
	Call(context.Context, UpdateMetricsServiceParams) (UpdateMetricsServiceResult, error)
}

var _ IUpdateMetricsService = (*UpdateMetricsService)(nil)
var _ CntrRep = (*repositories.CounterRepository)(nil)
var _ GGRep = (*repositories.GaugeRepository)(nil)

// UpdateMetricsService это сервис, который отвечает за логику обновления/создания сразу нескольких метрик.
type UpdateMetricsService struct {
	counterRep CntrRep
	gaugeRep   GGRep
	lg         *logging.ZapLogger
}

func NewUpdateMetricsService(counterRep CntrRep, gaugeRep GGRep, lg *logging.ZapLogger) *UpdateMetricsService {
	return &UpdateMetricsService{
		counterRep: counterRep,
		gaugeRep:   gaugeRep,
		lg:         lg,
	}
}

type UpdateMetricsServiceEl struct {
	ID    string  `json:"id"`
	MType string  `json:"type"`
	Delta int64   `json:"delta,omitempty"`
	Value float64 `json:"value,omitempty"`
}

// UpdateMetricsServiceParams параметры, которые необходимо передать в метод Call, для выполнения логики сервиса UpdateMetricsService.
type UpdateMetricsServiceParams []UpdateMetricsServiceEl

// UpdateMetricsServiceResult рузультат работы сервиса UpdateMetricsService.
type UpdateMetricsServiceResult struct {
}

// Call принимает параметры и отвечает за выполнение логики сервиса UpdateMetricsService.
func (s *UpdateMetricsService) Call(ctx context.Context, params UpdateMetricsServiceParams) (UpdateMetricsServiceResult, error) {
	svcCtx := s.lg.WithContextFields(ctx, zap.Any("actor", "update_metrics_service"))

	cntrs, ggs := s.group(params)

	if len(cntrs) > 0 {
		if err := s.counterRep.SaveCollection(svcCtx, cntrs); err != nil {
			return UpdateMetricsServiceResult{}, fmt.Errorf("update_metrics_service.go: save counters error %w", err)
		}
	}

	if len(ggs) > 0 {
		if err := s.gaugeRep.SaveCollection(svcCtx, ggs); err != nil {
			return UpdateMetricsServiceResult{}, fmt.Errorf("update_metrics_service.go: save gauges error %w", err)
		}
	}

	return UpdateMetricsServiceResult{}, nil
}

func (s *UpdateMetricsService) group(total UpdateMetricsServiceParams) (cntrs []models.Counter, gs []models.Gauge) {
	cntrs = make([]models.Counter, 0, len(total))
	gs = make([]models.Gauge, 0, len(total))

	for _, el := range total {
		if el.MType == models.CounterType {
			cntrs = append(cntrs, s.buildCounter(el))
		} else {
			gs = append(gs, s.buildGauge(el))
		}
	}
	return
}

func (s *UpdateMetricsService) buildCounter(el UpdateMetricsServiceEl) models.Counter {
	return models.Counter{
		Name:  el.ID,
		Value: el.Delta,
	}
}

func (s *UpdateMetricsService) buildGauge(el UpdateMetricsServiceEl) models.Gauge {
	return models.Gauge{
		Name:  el.ID,
		Value: el.Value,
	}
}
