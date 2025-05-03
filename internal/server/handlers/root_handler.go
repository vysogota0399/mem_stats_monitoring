package handlers

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
)

type RootCounterRepository interface {
	All(ctx context.Context) ([]models.Counter, error)
}

type RootGaugeRepository interface {
	All(ctx context.Context) ([]models.Gauge, error)
}

var _ RootCounterRepository = (*repositories.CounterRepository)(nil)
var _ RootGaugeRepository = (*repositories.GaugeRepository)(nil)

type RootHandler struct {
	counter RootCounterRepository
	gauge   RootGaugeRepository
}

func NewRootHandler(counter RootCounterRepository, gauge RootGaugeRepository) *RootHandler {
	return &RootHandler{
		counter: counter,
		gauge:   gauge,
	}
}

//go:embed templates
var rootHandlerTemplates embed.FS

func (h *RootHandler) Registrate() (server.Route, error) {
	tmp := template.New("root")
	if _, err := tmp.ParseFS(rootHandlerTemplates, "templates/root/index.tmpl"); err != nil {
		return server.Route{}, fmt.Errorf("root_handler: parse fs error %w", err)
	}

	return server.Route{
		Path:          "/",
		Method:        "GET",
		Handler:       h.handler(),
		HTMLTemplates: []*template.Template{tmp},
	}, nil
}

func (h *RootHandler) handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		gauge, err := h.gauge.All(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		counter, err := h.counter.All(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"gauge":   gauge,
			"counter": counter,
		})
	}
}
