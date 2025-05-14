package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type IUpdateMetricService interface {
	Call(ctx context.Context, params service.UpdateMetricServiceParams) (service.UpdateMetricServiceResult, error)
}

var _ IUpdateMetricService = (*service.UpdateMetricService)(nil)

type UpdateMetricHandler struct {
	service IUpdateMetricService
	lg      *logging.ZapLogger
}

func NewUpdateMetricHandler(service IUpdateMetricService, lg *logging.ZapLogger) *UpdateMetricHandler {
	return &UpdateMetricHandler{
		service: service,
		lg:      lg,
	}
}

func (h *UpdateMetricHandler) Registrate() (server.Route, error) {
	return server.Route{
		Path:    "/update/:type/:name/:value",
		Method:  "POST",
		Handler: h.handler(),
	}, nil
}

type Metric struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (h *UpdateMetricHandler) handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := utils.InitHandlerCtx(c, h.lg, "update_metric_handler")
		mType := c.Param("type")
		mName := c.Param("name")
		mValue := c.Param("value")

		if mType != models.GaugeType && mType != models.CounterType {
			h.lg.ErrorCtx(ctx, "invalid metric type", zap.String("type", mType))
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if mName == "" {
			h.lg.ErrorCtx(ctx, "invalid metric name", zap.String("name", mName))
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		var delta *int64
		var value *float64

		if mType == models.GaugeType {
			v, err := strconv.ParseFloat(mValue, 64)
			if err != nil {
				h.lg.ErrorCtx(ctx, "invalid metric value", zap.Error(err))
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			value = &v
		} else {
			d, err := strconv.ParseInt(mValue, 10, 64)
			if err != nil {
				h.lg.ErrorCtx(ctx, "invalid metric value", zap.Error(err))
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			delta = &d
		}

		params := service.UpdateMetricServiceParams{
			MName: mName,
			MType: mType,
		}

		if delta != nil {
			params.Delta = *delta
		}

		if value != nil {
			params.Value = *value
		}

		if _, err := h.service.Call(c, params); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
		}
	}
}
