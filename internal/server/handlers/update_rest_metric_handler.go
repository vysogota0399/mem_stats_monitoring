package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

// Metrics передается в payload запроса.
type Metrics struct {
	ID    string  `json:"id" binding:"required"`   // имя метрики.
	MType string  `json:"type" binding:"required"` // параметр, принимающий значение gauge или counter.
	Delta int64   `json:"delta,omitempty"`         // значение метрики в случае передачи counter.
	Value float64 `json:"value,omitempty"`         // значение метрики в случае передачи gauge.
}

type IUpdateRestMetricService interface {
	Call(context.Context, service.UpdateMetricServiceParams) (service.UpdateMetricServiceResult, error)
}

var _ IUpdateRestMetricService = (*service.UpdateMetricService)(nil)

// UpdateRestMetricHandler обработчик позволяет сохранить произвольную метрику.
type UpdateRestMetricHandler struct {
	service IUpdateRestMetricService
	lg      *logging.ZapLogger
}

func NewUpdateRestMetricHandler(srvc IUpdateRestMetricService, lg *logging.ZapLogger) *UpdateRestMetricHandler {
	return &UpdateRestMetricHandler{
		service: srvc,
		lg:      lg,
	}
}

func (h *UpdateRestMetricHandler) Registrate() (server.Route, error) {
	return server.Route{
		Path:    "/update/",
		Method:  "POST",
		Handler: h.handler(),
	}, nil
}

func (h *UpdateRestMetricHandler) handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := utils.InitHandlerCtx(c, h.lg, "update_rest_metrics_handler")

		var metric Metrics
		if err := c.ShouldBindJSON(&metric); err != nil {
			h.lg.DebugCtx(ctx, "invalid params", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		params := service.UpdateMetricServiceParams{
			MName: metric.ID,
			MType: metric.MType,
		}

		h.lg.DebugCtx(ctx, "update metric", zap.Any("metric", metric))

		if metric.MType == models.GaugeType {
			params.Value = metric.Value
		} else {
			params.Delta = metric.Delta
		}

		result, err := h.service.Call(ctx, params)
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
