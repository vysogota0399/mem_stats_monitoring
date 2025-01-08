package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

const gauge string = "gauge"
const counter string = "counter"

type ShowMetricHandler struct {
	gaugeRepository   repositories.Gauge
	counterRepository repositories.Counter
}

func NewShowMetricHandler(strg storage.Storage) gin.HandlerFunc {
	return showMetricHandlerFunc(
		&ShowMetricHandler{
			gaugeRepository:   repositories.NewGauge(strg),
			counterRepository: repositories.NewCounter(strg),
		},
	)
}

func showMetricHandlerFunc(h *ShowMetricHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		record, err := h.fetchMetic(c.Param("type"), c.Param("name"))

		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		result := record.StringValue()
		c.String(http.StatusOK, result)
	}
}

func (h *ShowMetricHandler) fetchMetic(mType, mName string) (models.Metricable, error) {
	switch mType {
	case gauge:
		return h.gaugeRepository.Last(mName)
	case counter:
		return h.counterRepository.Last(mName)
	}
	return nil, fmt.Errorf("internal/server/handlers/show_metric_handler.go: fatch mType: %s, mName: %s error", mType, mName)
}
