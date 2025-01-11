package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type UpdatesRestMetricHandler struct {
	storage storage.Storage
	service UpdateMetricService
	lg      *logging.ZapLogger
}

func NewUpdatesRestMetricHandler(s storage.Storage, srvc *service.Service, lg *logging.ZapLogger) gin.HandlerFunc {
	return updatesRestMetricHandlerFunc(
		&UpdatesRestMetricHandler{
			storage: s,
			service: srvc.UpdateMetricService,
			lg:      lg,
		},
	)
}

type updatesReqSchema struct {
	Metrics []Metrics
}

func updatesRestMetricHandlerFunc(h *UpdatesRestMetricHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqSchema updatesReqSchema
		ctx := utils.InitHandlerCtx(c, h.lg, "updates_rest_metrics_handler")
		c.Writer.Header().Add("Content-Type", "application/json")

		if err := c.ShouldBindJSON(&reqSchema); err != nil {
			h.lg.DebugCtx(ctx, "invalid params", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		
	}
}
