package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

const gauge string = "gauge"
const counter string = "counter"

type ShowMetricHandler struct {
	gaugeRepository   *repositories.Gauge
	counterRepository *repositories.Counter
	lg                *logging.ZapLogger
}

func NewShowMetricHandler(strg storage.Storage, lg *logging.ZapLogger) gin.HandlerFunc {
	return showMetricHandlerFunc(
		&ShowMetricHandler{
			gaugeRepository:   repositories.NewGauge(strg, lg),
			counterRepository: repositories.NewCounter(strg, lg),
			lg:                lg,
		},
	)
}

func showMetricHandlerFunc(h *ShowMetricHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		record, err := h.fetchMetic(c, c.Param("type"), c.Param("name"))

		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		result := record.StringValue()
		c.String(http.StatusOK, result)
	}
}

func (h *ShowMetricHandler) fetchMetic(c context.Context, mType, mName string) (models.Metricable, error) {
	switch mType {
	case gauge:
		return h.gaugeRepository.Last(c, mName)
	case counter:
		return h.counterRepository.Last(c, mName)
	}
	return nil, fmt.Errorf("internal/server/handlers/show_metric_handler.go: fatch mType: %s, mName: %s error", mType, mName)
}
