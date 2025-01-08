package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

type metricsUpdater func(m Metric, storage storage.Storage) error

type UpdateMetricHandler struct {
	storage        storage.Storage
	metricsUpdater metricsUpdater
}

func NewUpdateMetricHandler(strg storage.Storage) gin.HandlerFunc {
	return updateMetricHandlerFunc(
		&UpdateMetricHandler{
			storage:        strg,
			metricsUpdater: updateMetrics,
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

		if err := h.metricsUpdater(metric, h.storage); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
		}
	}
}

func updateMetrics(m Metric, strg storage.Storage) error {
	if m.Type != "gauge" && m.Type != "counter" {
		return fmt.Errorf("update_metric_service: underfined metric type: %s", m.Type)
	}

	if m.Type == "gauge" {
		err := processGauge(&m, strg)
		if err != nil {
			return err
		}
	} else {
		err := processCounter(&m, strg)
		if err != nil {
			return err
		}
	}

	return nil
}

func processGauge(m *Metric, strg storage.Storage) error {
	g, err := newGauge(m)
	if err != nil {
		return err
	}

	rep := repositories.NewGauge(strg)
	if _, err := rep.Craete(g); err != nil {
		return err
	}

	return nil
}

func processCounter(m *Metric, strg storage.Storage) error {
	c, err := newCounter(m)
	if err != nil {
		return err
	}

	rep := repositories.NewCounter(strg)
	if _, err := rep.Craete(c); err != nil {
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
