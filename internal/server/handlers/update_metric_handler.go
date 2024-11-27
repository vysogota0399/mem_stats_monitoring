package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

const updateMeticsContentType string = "text/plain"

type metricsUpdater func(m Metric, storage storage.Storage, logger utils.Logger) error

type UpdateMetricHandler struct {
	logger         utils.Logger
	storage        storage.Storage
	metricsUpdater metricsUpdater
}

func NewUpdateMetricHandler(storage storage.Storage, logger utils.Logger) *UpdateMetricHandler {
	return &UpdateMetricHandler{
		logger:         logger,
		storage:        storage,
		metricsUpdater: updateMetrics,
	}
}

type Metric struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (m Metric) String() string {
	mJSON, err := json.Marshal(m)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintln(string(mJSON))
}

func (h UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	if err := validateHeader(r); err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metric, err := parsePath(r.URL.Path)
	if err != nil {
		h.logger.Println(err)
		http.Error(w, "", http.StatusNotFound)
		return
	}
	h.logger.Printf("request params: %v", metric)

	if err := h.metricsUpdater(metric, h.storage, h.logger); err != nil {
		h.logger.Printf("Update metric error\n%v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func parsePath(path string) (Metric, error) {
	pattern := regexp.MustCompile(`^/update/[a-zA-Z]+/[a-zA-Z]+/[a-zA-Z0-9]+$`)
	if !pattern.MatchString(path) {
		return Metric{}, errors.New("update_metric_handler: route not found")
	}

	params := strings.Split(path, "/")
	params = params[len(params)-3:]
	if len(params) != 3 {
		return Metric{}, errors.New("update_metric_handler: route not found")
	}

	return Metric{Type: params[0], Name: params[1], Value: params[2]}, nil
}

func validateHeader(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain" {
		return fmt.Errorf("update_metric_handler: expected content type: %s, got: %s", updateMeticsContentType, contentType)
	}

	return nil
}

func updateMetrics(m Metric, storage storage.Storage, logger utils.Logger) error {
	if m.Type != "gauge" && m.Type != "counter" {
		return fmt.Errorf("update_metric_service: underfined metric type: %s", m.Type)
	}

	if m.Type == "gauge" {
		err := processGauge(&m, storage, logger)
		if err != nil {
			return err
		}
	} else {
		err := processCounter(&m, storage, logger)
		if err != nil {
			return err
		}
	}

	return nil
}

func processGauge(m *Metric, storage storage.Storage, logger utils.Logger) error {
	g, err := newGauge(m)
	if err != nil {
		return err
	}

	logger.Printf("New gauge: %v", g)
	rep := repositories.NewGauge(storage)
	if _, err := rep.Craete(g); err != nil {
		return err
	}

	return nil
}

func processCounter(m *Metric, storage storage.Storage, logger utils.Logger) error {
	c, err := newCounter(m)
	if err != nil {
		return err
	}

	logger.Printf("New counter: %v", c)
	rep := repositories.NewCounter(storage)
	if _, err := rep.Craete(c); err != nil {
		return err
	}

	return nil
}

func newGauge(m *Metric) (models.Gauge, error) {
	value, err := strconv.ParseFloat(m.Value, 64)
	if err != nil {
		return models.Gauge{}, fmt.Errorf("update_metric_service: %v is not float64", m.Value)
	}

	g := models.Gauge{
		Name:  m.Name,
		Value: value,
	}

	return g, nil
}

func newCounter(m *Metric) (models.Counter, error) {
	value, err := strconv.ParseInt(m.Value, 10, 64)
	if err != nil {
		return models.Counter{}, fmt.Errorf("update_metric_service: %v is not int64", m.Value)
	}

	c := models.Counter{
		Name:  m.Name,
		Value: value,
	}

	return c, nil
}
