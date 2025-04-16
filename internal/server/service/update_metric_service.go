package service

import (
	"context"
	"fmt"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
)

type CreateCntrRep interface {
	Create(context.Context, *models.Counter) error
}

type CreateGaugeRep interface {
	Create(context.Context, *models.Gauge) error
}

// IUpdateMetricService interface for mocks
type IUpdateMetricService interface {
	Call(context.Context, UpdateMetricServiceParams) (UpdateMetricServiceResult, error)
}

var _ IUpdateMetricService = (*UpdateMetricService)(nil)
var _ CreateCntrRep = (*repositories.CounterRepository)(nil)
var _ CreateGaugeRep = (*repositories.GaugeRepository)(nil)

// UpdateMetricService это сервис, который отвечает за логику обновления/создания одной метрики.
type UpdateMetricService struct {
	counterRep CreateCntrRep
	gaugeRep   CreateGaugeRep
}

func NewUpdateMetricService(counterRep CreateCntrRep, gaugeRep CreateGaugeRep) *UpdateMetricService {
	return &UpdateMetricService{
		counterRep: counterRep,
		gaugeRep:   gaugeRep,
	}
}

// UpdateMetricServiceParams параметры, которые необходимо передать в метод Call, для выполнения логики сервиса UpdateMetricService.
type UpdateMetricServiceParams struct {
	MName string
	MType string
	Delta int64
	Value float64
}

// UpdateMetricServiceResult рузультат работы сервиса UpdateMetricService.
type UpdateMetricServiceResult struct {
	ID    string  `json:"id"`
	MType string  `json:"type"`
	Delta int64   `json:"delta,omitempty"`
	Value float64 `json:"value,omitempty"`
}

// Call принимает параметры и отвечает за выполнение логики сервиса UpdateMetricService.
func (s UpdateMetricService) Call(ctx context.Context, params UpdateMetricServiceParams) (UpdateMetricServiceResult, error) {
	switch params.MType {
	case models.CounterType:
		return s.createCounter(ctx, params)
	case models.GaugeType:
		return s.createGauge(ctx, params)
	default:
		return UpdateMetricServiceResult{}, fmt.Errorf("internal/server/service/update_metric_service.go: unexpected type(%s)", params.MType)
	}
}

func (s UpdateMetricService) createCounter(ctx context.Context, params UpdateMetricServiceParams) (UpdateMetricServiceResult, error) {
	cntr := models.Counter{
		Name:  params.MName,
		Value: params.Delta,
	}

	if err := s.counterRep.Create(ctx, &cntr); err != nil {
		return UpdateMetricServiceResult{}, fmt.Errorf("internal/server/service/update_metric_service.go: create counter %+v error %w", cntr, err)
	}

	return UpdateMetricServiceResult{
		ID:    params.MName,
		MType: params.MType,
		Delta: cntr.Value,
	}, nil
}

func (s UpdateMetricService) createGauge(ctx context.Context, params UpdateMetricServiceParams) (UpdateMetricServiceResult, error) {
	gauge := models.Gauge{
		Name:  params.MName,
		Value: params.Value,
	}

	if err := s.gaugeRep.Create(ctx, &gauge); err != nil {
		return UpdateMetricServiceResult{}, fmt.Errorf("internal/server/service/update_metric_service.go: create gauge error %w", err)
	}

	return UpdateMetricServiceResult{
		ID:    params.MName,
		MType: params.MType,
		Value: gauge.Value,
	}, nil
}
