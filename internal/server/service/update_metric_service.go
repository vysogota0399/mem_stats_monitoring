package service

import (
	"context"
	"fmt"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
)

// UpdateMetricService это сервис, который отвечает за логику обновления/создания одной метрики.
type UpdateMetricService struct {
	counterRep *repositories.Counter
	gaugeRep   *repositories.Gauge
}

// UpdateMetricServiceParams параметры, которые необходимо передать в метод Call, для выполнения логики сервиса UpdateMetricService.
type UpdateMetricServiceParams struct {
	MName string
	MType string
	Delta *int64
	Value *float64
}

// UpdateMetricServiceResult рузультат работы сервиса UpdateMetricService.
type UpdateMetricServiceResult struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
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
	var value int64
	if params.Delta != nil {
		value = *params.Delta
	}

	record, err := s.counterRep.Create(
		ctx,
		&models.Counter{
			Name:  params.MName,
			Value: value,
		})

	if err != nil {
		return UpdateMetricServiceResult{}, err
	}

	return UpdateMetricServiceResult{
		ID:    params.MName,
		MType: params.MType,
		Delta: &record.Value,
	}, nil
}

func (s UpdateMetricService) createGauge(ctx context.Context, params UpdateMetricServiceParams) (UpdateMetricServiceResult, error) {
	var value float64
	if params.Value != nil {
		value = *params.Value
	}

	record, err := s.gaugeRep.Create(
		ctx,
		&models.Gauge{
			Name:  params.MName,
			Value: value,
		})

	if err != nil {
		return UpdateMetricServiceResult{}, err
	}

	return UpdateMetricServiceResult{
		ID:    params.MName,
		MType: params.MType,
		Value: &record.Value,
	}, nil
}
