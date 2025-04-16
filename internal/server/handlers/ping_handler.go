package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type PingHandler struct {
	strg storages.Storage
	lg   *logging.ZapLogger
}

func (h *PingHandler) Registrate() (server.Route, error) {
	if _, ok := h.strg.(*storages.PG); !ok {
		return server.Route{}, nil
	}

	return server.Route{
		Path:    "/ping",
		Method:  "GET",
		Handler: h.handler(),
	}, nil
}

func NewPingHandler(strg storages.Storage, lg *logging.ZapLogger) *PingHandler {
	return &PingHandler{strg: strg, lg: lg}
}

func (h *PingHandler) handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := utils.InitHandlerCtx(c, h.lg, "ping_handler")

		if err := h.strg.Ping(ctx); err != nil {
			h.lg.ErrorCtx(ctx, "ping failed", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusOK)
	}
}
