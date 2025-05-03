package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type IUpdateMetricsService interface {
	Call(context.Context, service.UpdateMetricsServiceParams) (service.UpdateMetricsServiceResult, error)
}

var _ IUpdateMetricsService = (*service.UpdateMetricsService)(nil)

// UpdatesRestMetricsHandler обработчик позволяет сохранить пачку произвольных метрик.
type UpdatesRestMetricsHandler struct {
	service IUpdateMetricsService
	lg      *logging.ZapLogger
}

func NewUpdatesRestMetricsHandler(srvc IUpdateMetricsService, lg *logging.ZapLogger) *UpdatesRestMetricsHandler {
	return &UpdatesRestMetricsHandler{
		service: srvc,
		lg:      lg,
	}
}

func (h *UpdatesRestMetricsHandler) Registrate() (server.Route, error) {
	return server.Route{
		Path:    "/updates/",
		Method:  "POST",
		Handler: h.handler(),
	}, nil
}

func (h *UpdatesRestMetricsHandler) handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var params service.UpdateMetricsServiceParams
		ctx := utils.InitHandlerCtx(c, h.lg, "updates_rest_metrics_handler")

		if err := c.ShouldBindJSON(&params); err != nil {
			h.lg.DebugCtx(ctx, "invalid params", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		if len(params) == 0 {
			h.lg.DebugCtx(ctx, "empty metrics array")
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
