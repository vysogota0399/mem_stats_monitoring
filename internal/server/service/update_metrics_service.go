package service

import (
	"context"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
)

type CntrRep interface {
	SaveCollection(context.Context, []models.Counter) ([]models.Counter, error)
}

type GGRep interface {
	SaveCollection(context.Context, []models.Gauge) ([]models.Gauge, error)
}

// UpdateMetricsService это сервис, который отвечает за логику обновления/создания сразу нескольких метрик.
type UpdateMetricsService struct {
	counterRep CntrRep
	gaugeRep   GGRep
}

type UpdateMetricsServiceEl struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

// UpdateMetricsServiceParams параметры, которые необходимо передать в метод Call, для выполнения логики сервиса UpdateMetricsService.
type UpdateMetricsServiceParams []UpdateMetricsServiceEl

// UpdateMetricsServiceResult рузультат работы сервиса UpdateMetricsService.

type UpdateMetricsServiceResult struct {
	cntrs []models.Counter
	ggs   []models.Gauge
}

// Call принимает параметры и отвечает за выполнение логики сервиса UpdateMetricsService.
func (s *UpdateMetricsService) Call(ctx context.Context, params UpdateMetricsServiceParams) (*UpdateMetricsServiceResult, error) {
	cntrs, ggs := s.group(params)

	cntrs, err := s.counterRep.SaveCollection(ctx, cntrs)
	if err != nil {
		return nil, err
	}

	ggs, err = s.gaugeRep.SaveCollection(ctx, ggs)
	if err != nil {
		return nil, err
	}

	return &UpdateMetricsServiceResult{
		ggs:   ggs,
		cntrs: cntrs,
	}, nil
}

func (s *UpdateMetricsService) group(total UpdateMetricsServiceParams) (cntrs []models.Counter, gs []models.Gauge) {
	cntrs = make([]models.Counter, 0)
	gs = make([]models.Gauge, 0)

	for _, el := range total {
		cntrs, gs = s.appendEl(el, cntrs, gs)
	}
	return
}

func (s *UpdateMetricsService) appendEl(
	el UpdateMetricsServiceEl,
	cntrs []models.Counter,
	gs []models.Gauge) ([]models.Counter, []models.Gauge) {
	if el.MType == models.CounterType {
		cntrs = append(cntrs, s.buildCounter(el))
	} else {
		gs = append(gs, s.buildGauge(el))
	}

	return cntrs, gs
}

func (s *UpdateMetricsService) buildCounter(el UpdateMetricsServiceEl) models.Counter {
	var value int64
	if el.Delta != nil {
		value = *el.Delta
	}

	return models.Counter{
		Name:  el.ID,
		Value: value,
	}
}

func (s *UpdateMetricsService) buildGauge(el UpdateMetricsServiceEl) models.Gauge {
	var value float64
	if el.Value != nil {
		value = *el.Value
	}

	return models.Gauge{
		Name:  el.ID,
		Value: value,
	}
}
