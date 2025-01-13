package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type ShowRestMetricHandler struct {
	gaugeRepository   *repositories.Gauge
	counterRepository *repositories.Counter
	lg                *logging.ZapLogger
}

func NewShowRestMetricHandler(strg storage.Storage, lg *logging.ZapLogger) gin.HandlerFunc {
	return showRestMetricHandlerFunc(
		&ShowRestMetricHandler{
			gaugeRepository:   repositories.NewGauge(strg),
			counterRepository: repositories.NewCounter(strg),
			lg:                lg,
		},
	)
}

type showRestMetricParams struct {
	ID    string `json:"id" bind:"required"`   // имя метрики
	MType string `json:"type" bind:"required"` // параметр, принимающий значение gauge или counter
}

type showRestMetricResponse struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func showRestMetricHandlerFunc(h *ShowRestMetricHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := utils.InitHandlerCtx(c, h.lg, "show_rest_metrics_handler")
		c.Writer.Header().Add("Content-Type", "application/json")
		var params showRestMetricParams

		if err := c.ShouldBindJSON(&params); err != nil {
			h.lg.DebugCtx(ctx, "Invalid params", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		response, err := h.fetchMetic(c, params.MType, params.ID)
		if err != nil {
			if errors.Is(err, storage.ErrNoRecords) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{})
				return
			}
			h.lg.ErrorCtx(ctx, "Show metric error", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		if err := json.NewEncoder(c.Writer).Encode(response); err != nil {
			h.lg.ErrorCtx(ctx, "Encoding response error", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}
}

func (h *ShowRestMetricHandler) fetchMetic(ctx context.Context, mType, mName string) (showRestMetricResponse, error) {
	switch mType {
	case gauge:
		record, err := h.gaugeRepository.Last(ctx, mName)
		if err != nil {
			return showRestMetricResponse{}, err
		}

		return showRestMetricResponse{
			ID:    record.Name,
			MType: models.GaugeType,
			Value: &record.Value,
		}, nil
	case counter:
		record, err := h.counterRepository.Last(ctx, mName)
		if err != nil {
			return showRestMetricResponse{}, err
		}

		return showRestMetricResponse{
			ID:    record.Name,
			MType: models.CounterType,
			Delta: &record.Value,
		}, nil
	}

	return showRestMetricResponse{}, fmt.Errorf(
		"internal/server/handlers/show_metric_handler.go: fatch mType: %s, mName: %s error",
		mType,
		mName,
	)
}
