package service

import (
	"fmt"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
)

type UpdateMetricService struct {
	counterRep repositories.Counter
	gaugeRep   repositories.Gauge
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

func (s UpdateMetricService) Call(params UpdateMetricServiceParams) (UpdateMetricServiceResult, error) {
	if params.MType == "counter" && params.Delta != nil {
		return s.createCounter(params)
	} else if params.MType == "gauge" && params.Value != nil {
		return s.createGauge(params)
	}

	return UpdateMetricServiceResult{}, fmt.Errorf(
		"internal/server/service/update_metric_service.go: unexpected type(%s) or delta(%v)/value(%v)",
		params.MType,
		params.Delta,
		params.Value,
	)
}

func (s UpdateMetricService) createCounter(params UpdateMetricServiceParams) (UpdateMetricServiceResult, error) {
	record, err := s.counterRep.Craete(models.Counter{
		Name:  params.MName,
		Value: *params.Delta,
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

func (s UpdateMetricService) createGauge(params UpdateMetricServiceParams) (UpdateMetricServiceResult, error) {
	record, err := s.gaugeRep.Craete(models.Gauge{
		Name:  params.MName,
		Value: *params.Value,
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
