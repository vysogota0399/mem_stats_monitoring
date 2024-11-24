package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

const updateMeticsContentType string = "text/plain"

type UpdateMetricHandler struct {
	logger utils.Logger
	app    *Server
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

	if err := UpdateMetricService(metric, h.app.storage, h.logger); err != nil {
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
