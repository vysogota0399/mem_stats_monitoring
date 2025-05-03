package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type IShowRestMetricGaugeRepository interface {
	FindByName(ctx context.Context, name string) (models.Gauge, error)
}

type IShowRestMetricCounterRepository interface {
	FindByName(ctx context.Context, name string) (models.Counter, error)
}

var _ IShowRestMetricGaugeRepository = (*repositories.GaugeRepository)(nil)
var _ IShowRestMetricCounterRepository = (*repositories.CounterRepository)(nil)

type ShowRestMetricHandler struct {
	gaugeRepository   IShowRestMetricGaugeRepository
	counterRepository IShowRestMetricCounterRepository
	lg                *logging.ZapLogger
}

func NewShowRestMetricHandler(gaugeRepository IShowRestMetricGaugeRepository, counterRepository IShowRestMetricCounterRepository, lg *logging.ZapLogger) *ShowRestMetricHandler {
	return &ShowRestMetricHandler{
		gaugeRepository:   gaugeRepository,
		counterRepository: counterRepository,
		lg:                lg,
	}
}

func (h *ShowRestMetricHandler) Registrate() (server.Route, error) {
	return server.Route{
		Path:    "/value/",
		Method:  "GET",
		Handler: h.handler(),
	}, nil
}

type showRestMetricParams struct {
	ID    string `json:"id" bind:"required"`   // имя метрики
	MType string `json:"type" bind:"required"` // параметр, принимающий значение gauge или counter
}

type showRestMetricResponse struct {
	ID    string  `json:"id"`              // имя метрики
	MType string  `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (h *ShowRestMetricHandler) handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := utils.InitHandlerCtx(c, h.lg, "show_rest_metrics_handler")

		var params showRestMetricParams

		if err := c.ShouldBindJSON(&params); err != nil {
			h.lg.DebugCtx(ctx, "invalid params", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		response, err := h.fetchMetic(c, params.MType, params.ID)
		if err != nil {
			if errors.Is(err, storages.ErrNoRecords) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{})
				return
			}
			h.lg.ErrorCtx(ctx, "show metric error", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		if err := json.NewEncoder(c.Writer).Encode(response); err != nil {
			h.lg.ErrorCtx(ctx, "encoding response error", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}
}

func (h *ShowRestMetricHandler) fetchMetic(ctx context.Context, mType, mName string) (showRestMetricResponse, error) {
	switch mType {
	case models.GaugeType:
		record, err := h.gaugeRepository.FindByName(ctx, mName)
		if err != nil {
			return showRestMetricResponse{}, err
		}

		return showRestMetricResponse{
			ID:    record.Name,
			MType: models.GaugeType,
			Value: record.Value,
		}, nil
	case models.CounterType:
		record, err := h.counterRepository.FindByName(ctx, mName)
		if err != nil {
			return showRestMetricResponse{}, err
		}

		return showRestMetricResponse{
			ID:    record.Name,
			MType: models.CounterType,
			Delta: record.Value,
		}, nil
	}

	return showRestMetricResponse{}, fmt.Errorf(
		"internal/server/handlers/show_rest_metric_handler.go: fetch metric error mType: %s, mName: %s",
		mType,
		mName,
	)
}
