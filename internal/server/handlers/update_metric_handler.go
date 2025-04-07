package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

type metricsUpdater func(ctx context.Context, m Metric, storage storage.Storage, lg *logging.ZapLogger) error

type UpdateMetricHandler struct {
	storage        storage.Storage
	metricsUpdater metricsUpdater
	lg             *logging.ZapLogger
}

func NewUpdateMetricHandler(strg storage.Storage, lg *logging.ZapLogger) gin.HandlerFunc {
	return updateMetricHandlerFunc(
		&UpdateMetricHandler{
			storage:        strg,
			metricsUpdater: updateMetrics,
			lg:             lg,
		},
	)
}

type Metric struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (m Metric) String() string {
	mJSON, err := json.Marshal(m)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintln(string(mJSON))
}

func updateMetricHandlerFunc(h *UpdateMetricHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		metric := Metric{
			Name:  c.Param("name"),
			Type:  c.Param("type"),
			Value: c.Param("value"),
		}

		if err := h.metricsUpdater(c, metric, h.storage, h.lg); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
		}
	}
}

func updateMetrics(ctx context.Context, m Metric, strg storage.Storage, lg *logging.ZapLogger) error {
	if m.Type != models.GaugeType && m.Type != models.CounterType {
		return fmt.Errorf("update_metric_service: underfined metric type: %s", m.Type)
	}

	if m.Type == "gauge" {
		err := processGauge(ctx, &m, strg, lg)
		if err != nil {
			return err
		}
	} else {
		err := processCounter(ctx, &m, strg, lg)
		if err != nil {
			return err
		}
	}

	return nil
}

func processGauge(ctx context.Context, m *Metric, strg storage.Storage, lg *logging.ZapLogger) error {
	g, err := newGauge(m)
	if err != nil {
		return err
	}

	rep := repositories.NewGauge(strg, lg)
	if _, err := rep.Create(ctx, &g); err != nil {
		return err
	}

	return nil
}

func processCounter(ctx context.Context, m *Metric, strg storage.Storage, lg *logging.ZapLogger) error {
	c, err := newCounter(m)
	if err != nil {
		return err
	}

	rep := repositories.NewCounter(strg, lg)
	if _, err := rep.Create(ctx, &c); err != nil {
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

func newCounter(m *Metric) (models.Counter, error) {
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
