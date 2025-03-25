package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type UpdateMetricsService interface {
	Call(context.Context, service.UpdateMetricsServiceParams) (*service.UpdateMetricsServiceResult, error)
}

// UpdatesRestMetricHandler обработчик позволяет сохранить пачку произвольных метрик.
type UpdatesRestMetricHandler struct {
	storage storage.Storage
	service UpdateMetricsService
	lg      *logging.ZapLogger
}

func NewUpdatesRestMetricHandler(s storage.Storage, srvc *service.Service, lg *logging.ZapLogger) gin.HandlerFunc {
	return updatesRestMetricHandlerFunc(
		&UpdatesRestMetricHandler{
			storage: s,
			service: srvc.UpdateMetricsService,
			lg:      lg,
		},
	)
}

func updatesRestMetricHandlerFunc(h *UpdatesRestMetricHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params service.UpdateMetricsServiceParams
		ctx := utils.InitHandlerCtx(c, h.lg, "updates_rest_metrics_handler")
		c.Writer.Header().Add("Content-Type", "application/json")

		if err := c.ShouldBindJSON(&params); err != nil {
			h.lg.DebugCtx(ctx, "invalid params", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		if _, err := h.service.Call(
			ctx,
			params,
		); err != nil {
			h.lg.ErrorCtx(ctx, "server error", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}
}
