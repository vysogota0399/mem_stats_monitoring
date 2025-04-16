package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type IShowMetricGaugeRepository interface {
	FindByName(ctx context.Context, name string) (models.Gauge, error)
}

type IShowMetricCounterRepository interface {
	FindByName(ctx context.Context, name string) (models.Counter, error)
}

var _ IShowMetricGaugeRepository = (*repositories.GaugeRepository)(nil)
var _ IShowMetricCounterRepository = (*repositories.CounterRepository)(nil)

type ShowMetricHandler struct {
	gaugeRepository   IShowMetricGaugeRepository
	counterRepository IShowMetricCounterRepository
	lg                *logging.ZapLogger
}

func NewShowMetricHandler(gaugeRepository IShowMetricGaugeRepository, counterRepository IShowMetricCounterRepository, lg *logging.ZapLogger) *ShowMetricHandler {
	return &ShowMetricHandler{
		gaugeRepository:   gaugeRepository,
		counterRepository: counterRepository,
		lg:                lg,
	}
}

func (h *ShowMetricHandler) Registrate() (server.Route, error) {
	return server.Route{
		Path:    "/value/:type/:name",
		Method:  "GET",
		Handler: h.handler(),
	}, nil
}

func (h *ShowMetricHandler) handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := utils.InitHandlerCtx(c, h.lg, "show_metric_handler")
		mType := c.Param("type")
		if mType != models.GaugeType && mType != models.CounterType {
			h.lg.ErrorCtx(ctx, "invalid metric type", zap.String("type", mType))
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		switch mType {
		case models.GaugeType:
			record, err := h.gaugeRepository.FindByName(ctx, c.Param("name"))
			if err != nil {
				if errors.Is(err, storages.ErrNoRecords) {
					c.AbortWithStatus(http.StatusNotFound)
					return
				}

				h.lg.ErrorCtx(ctx, "failed to find gauge by name", zap.Error(err))
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			c.String(http.StatusOK, record.StringValue())
		case models.CounterType:
			record, err := h.counterRepository.FindByName(ctx, c.Param("name"))
			if err != nil {
				if errors.Is(err, storages.ErrNoRecords) {
					c.AbortWithStatus(http.StatusNotFound)
					return
				}

				h.lg.ErrorCtx(ctx, "failed to find counter by name", zap.Error(err))
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			c.String(http.StatusOK, record.StringValue())
		}
	}
}
