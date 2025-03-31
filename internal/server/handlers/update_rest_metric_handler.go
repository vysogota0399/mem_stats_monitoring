package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

// Metrics передается в payload запроса.
type Metrics struct {
	ID    string   `json:"id" binding:"required"`   // имя метрики.
	MType string   `json:"type" binding:"required"` // параметр, принимающий значение gauge или counter.
	Delta *int64   `json:"delta,omitempty"`         // значение метрики в случае передачи counter.
	Value *float64 `json:"value,omitempty"`         // значение метрики в случае передачи gauge.
}

type UpdateMetricService interface {
	Call(context.Context, service.UpdateMetricServiceParams) (service.UpdateMetricServiceResult, error)
}

// UpdateRestMetricHandler обработчик позволяет сохранить произвольную метрику.
type UpdateRestMetricHandler struct {
	storage storage.Storage
	service UpdateMetricService
	lg      *logging.ZapLogger
}

func NewRestUpdateMetricHandler(s storage.Storage, srvc UpdateMetricService, lg *logging.ZapLogger) gin.HandlerFunc {
	return updateRestMetricHandlerFunc(
		&UpdateRestMetricHandler{
			storage: s,
			service: srvc,
			lg:      lg,
		},
	)
}

func updateRestMetricHandlerFunc(h *UpdateRestMetricHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := utils.InitHandlerCtx(c, h.lg, "update_rest_metrics_handler")
		c.Writer.Header().Add("Content-Type", "application/json")
		var metric Metrics
		if err := c.ShouldBindJSON(&metric); err != nil {
			h.lg.DebugCtx(ctx, "invalid params", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		result, err := h.service.Call(
			c,
			service.UpdateMetricServiceParams{
				MName: metric.ID,
				MType: metric.MType,
				Delta: metric.Delta,
				Value: metric.Value,
			},
		)

		if err != nil {
			h.lg.ErrorCtx(ctx, "update record failed", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		enc := json.NewEncoder(c.Writer)
		if err := enc.Encode(result); err != nil {
			h.lg.ErrorCtx(ctx, "error encoding response", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
}
