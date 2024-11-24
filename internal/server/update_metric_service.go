package server

import (
	"fmt"
	"strconv"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

func UpdateMetricService(m Metric, storage storage.Storage, logger utils.Logger) error {
	if m.Type != "gauge" && m.Type != "counter" {
		return fmt.Errorf("update_metric_service: underfined metric type: %s", m.Type)
	}

	if m.Type == "gauge" {
		err := processGauge(&m, storage, logger)
		if err != nil {
			return err
		}
	} else {
		err := processCounter(&m, storage, logger)
		if err != nil {
			return err
		}
	}

	return nil
}

func processGauge(m *Metric, storage storage.Storage, logger utils.Logger) error {
	g, err := newGauge(m)
	if err != nil {
		return err
	}

	logger.Printf("New gauge: %v", g)
	rep := repositories.NewGauge(storage)
	rep.Craete(g)
	return nil
}

func processCounter(m *Metric, storage storage.Storage, logger utils.Logger) error {
	c, err := NewCounter(m)
	if err != nil {
		return err
	}

	logger.Printf("New counter: %v", c)
	rep := repositories.NewCounter(storage)
	if err := rep.Craete(c); err != nil {
		return err
	}

	return nil
}

func newGauge(m *Metric) (models.Gauge, error) {
	value, err := strconv.ParseFloat(m.Value, 64)
	if err != nil {
		return models.Gauge{}, fmt.Errorf("update_metric_service: %v is not float64", m.Value)
	}

	g := models.Gauge{
		Name:  m.Name,
		Value: value,
	}

	return g, nil
}

func NewCounter(m *Metric) (models.Counter, error) {
	value, err := strconv.ParseInt(m.Value, 10, 64)
	if err != nil {
		return models.Counter{}, fmt.Errorf("update_metric_service: %v is not int64", m.Value)
	}

	c := models.Counter{
		Name:  m.Name,
		Value: value,
	}

	return c, nil
}
