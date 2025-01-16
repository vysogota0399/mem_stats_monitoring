package utils

import (
	"context"
	"math"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

func InitHandlerCtx(c *gin.Context, lg *logging.ZapLogger, handler string) context.Context {
	ctx := lg.WithContextFields(context.Background(),
		zap.String("name", handler),
	)
	rid, ok := c.Get("request_id")
	if !ok {
		return ctx
	}

	return lg.WithContextFields(ctx, zap.String("request_id", rid.(string)))
}

func Delay(i uint8) float64 {
	return math.Exp(float64(i) / 2)
}
