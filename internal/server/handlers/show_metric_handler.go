package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

type ShowMetricHandler struct {
	logger            utils.Logger
	gaugeRepository   repositories.Gauge
	counterRepository repositories.Counter
}

func NewShowMetricHandler(storage storage.Storage, logger utils.Logger) gin.HandlerFunc {
	return showMetricHandlerFunc(
		&ShowMetricHandler{
			logger:            logger,
			gaugeRepository:   repositories.NewGauge(storage),
			counterRepository: repositories.NewCounter(storage),
		},
	)
}

func showMetricHandlerFunc(h *ShowMetricHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		record, err := h.fetchMetic(c.Param("type"), c.Param("name"))

		if err != nil {
			h.logger.Println("Metric not found: %w", err)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		result := record.StringValue()
		h.logger.Printf("Metric found: %s", result)
		c.String(http.StatusOK, result)
	}
}

func (h *ShowMetricHandler) fetchMetic(mType, mName string) (models.Metricable, error) {
	switch mType {
	case "gauge":
		return h.gaugeRepository.Last(mName)
	case "counter":
		return h.counterRepository.Last(mName)
	}
	return nil, fmt.Errorf("show_metric_handler: fatch mType: %s, mName: %s error", mType, mName)
}
