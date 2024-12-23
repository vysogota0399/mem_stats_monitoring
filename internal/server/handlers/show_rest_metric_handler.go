package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/logger"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"go.uber.org/zap"
)

type ShowRestMetricHandler struct {
	gaugeRepository   repositories.Gauge
	counterRepository repositories.Counter
}

func NewShowRestMetricHandler(storage storage.Storage) gin.HandlerFunc {
	return showRestMetricHandlerFunc(
		&ShowRestMetricHandler{
			gaugeRepository:   repositories.NewGauge(storage),
			counterRepository: repositories.NewCounter(storage),
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
		c.Writer.Header().Add("Content-Type", "application/json")
		var params showRestMetricParams

		if err := c.ShouldBindJSON(&params); err != nil {
			logger.Log.Error(err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		response, err := h.fetchMetic(params.MType, params.ID)
		if err != nil {
			if errors.Is(err, storage.ErrNoRecords) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{})
				return
			}

			logger.Log.Error("show metric error", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
			return
		}

		logger.Log.Sugar().Infof("Response: %v", response)
		if err := json.NewEncoder(c.Writer).Encode(response); err != nil {
			logger.Log.Error("encoding response error", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}
}

func (h *ShowRestMetricHandler) fetchMetic(mType, mName string) (showRestMetricResponse, error) {
	switch mType {
	case gauge:
		record, err := h.gaugeRepository.Last(mName)
		if err != nil {
			return showRestMetricResponse{}, err
		}

		return showRestMetricResponse{
			ID:    record.Name,
			MType: models.GaugeType,
			Value: &record.Value,
		}, nil
	case counter:
		record, err := h.counterRepository.Last(mName)
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
