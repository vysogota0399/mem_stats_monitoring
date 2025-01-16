package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type PingHandler struct {
	storage storage.DBAble
	lg      *logging.ZapLogger
	skip    bool
}

func NewPingHandler(strg storage.Storage, lg *logging.ZapLogger) gin.HandlerFunc {
	h := &PingHandler{lg: lg}
	s, ok := strg.(storage.DBAble)
	if !ok {
		h.skip = true
	} else {
		h.storage = s
	}

	return PingHandlerFunc(h)
}

func PingHandlerFunc(h *PingHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := h.lg.WithContextFields(c, zap.String("name", "ping_handler"))
		if h.skip {
			h.lg.DebugCtx(ctx, "storage not implement database flow")
			c.AbortWithStatus(http.StatusNotAcceptable)
			return
		}

		if err := h.storage.Ping(); err != nil {
			h.lg.ErrorCtx(ctx, "ping failed", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}
}
