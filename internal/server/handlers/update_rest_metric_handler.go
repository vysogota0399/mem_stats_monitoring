package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/logger"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"go.uber.org/zap"
)

type Metrics struct {
	ID    string   `json:"id" binding:"required"`   // имя метрики
	MType string   `json:"type" binding:"required"` // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"`         // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"`         // значение метрики в случае передачи gauge
}

type UpdateMetricService interface {
	Call(service.UpdateMetricServiceParams) (service.UpdateMetricServiceResult, error)
}

type UpdateRestMetricHandler struct {
	storage storage.Storage
	service UpdateMetricService
}

func NewRestUpdateMetricHandler(s storage.Storage, service *service.Service) gin.HandlerFunc {
	return updateRestMetricHandlerFunc(
		&UpdateRestMetricHandler{
			storage: s,
			service: service.UpdateMetricService,
		},
	)
}

func updateRestMetricHandlerFunc(h *UpdateRestMetricHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Content-Type", "application/json")
		var metric Metrics
		if err := c.ShouldBindJSON(&metric); err != nil {
			logger.Log.Error(err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		result, err := h.service.Call(
			service.UpdateMetricServiceParams{
				MName: metric.ID,
				MType: metric.MType,
				Delta: metric.Delta,
				Value: metric.Value,
			},
		)

		if err != nil {
			logger.Log.Error("update record failed", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		enc := json.NewEncoder(c.Writer)
		if err := enc.Encode(result); err != nil {
			logger.Log.Error("error encoding response", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
}
