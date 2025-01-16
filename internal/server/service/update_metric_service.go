package service

import (
	"context"
	"fmt"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
)

type UpdateMetricService struct {
	counterRep *repositories.Counter
	gaugeRep   *repositories.Gauge
}

type UpdateMetricServiceParams struct {
	MName string
	MType string
	Delta *int64
	Value *float64
}

type UpdateMetricServiceResult struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (s UpdateMetricService) Call(ctx context.Context, params UpdateMetricServiceParams) (UpdateMetricServiceResult, error) {
	if params.MType == models.CounterType {
		return s.createCounter(ctx, params)
	} else if params.MType == models.GaugeType {
		return s.createGauge(ctx, params)
	}

	return UpdateMetricServiceResult{}, fmt.Errorf("internal/server/service/update_metric_service.go: unexpected type(%s)", params.MType)
}

func (s UpdateMetricService) createCounter(ctx context.Context, params UpdateMetricServiceParams) (UpdateMetricServiceResult, error) {
	var value int64
	if params.Delta != nil {
		value = *params.Delta
	}

	record, err := s.counterRep.Craete(
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

	record, err := s.gaugeRep.Craete(
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
