package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
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

func updatesRestMetricHandlerFunc(h *UpdatesRestMetricHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}
