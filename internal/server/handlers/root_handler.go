package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

type RootHandler struct {
	logger  utils.Logger
	storage storage.Storage
}

func NewRootHandler(storage storage.Storage, logger utils.Logger) gin.HandlerFunc {
	return RootHandlerFunc(
		&RootHandler{
			logger:  logger,
			storage: storage,
		},
	)
}

func RootHandlerFunc(h *RootHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		counterRep := repositories.NewCounter(h.storage)
		gaugeRep := repositories.NewGauge(h.storage)
		counterRecords := make([]models.Counter, 0)
		gaugeRecords := make([]models.Gauge, 0)

		for _, values := range counterRep.All() {
			count := len(values)
			if count == 0 {
				continue
			}

			counterRecords = append(counterRecords, values[count-1])
		}

		for _, values := range gaugeRep.All() {
			count := len(values)
			if count == 0 {
				continue
			}

			gaugeRecords = append(gaugeRecords, values[count-1])
		}

		h.logger.Printf("Gauge: %v\nCounter: %v", gaugeRecords, counterRecords)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"gauge":   gaugeRecords,
			"counter": counterRecords,
		})
	}
}
