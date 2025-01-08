package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

var ErrUnsuccessfulResponse error = errors.New("mem_stats_server: response not success")

type Reporter struct {
	client  *http.Client
	address string
	logger  utils.Logger
}

func NewReporter(address string) *Reporter {
	return &Reporter{address: address, client: &http.Client{}, logger: utils.InitLogger("[http]")}
}

func (c *Reporter) UpdateMetric(mType, mName, value string, requestID uuid.UUID) error {
	body, err := prepareBody(mType, mName, value)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		context.TODO(),
		"POST", fmt.Sprintf("%s/update/%s/%s/%v", c.address, mType, mName, value),
		body,
	)

	if err != nil {
		return fmt.Errorf("internal/agent/clients/reporter: create request err: %w", err)
	}

	req.Header.Add("Content-Type", "text/plain")
	req.Header.Add("X-Request-ID", requestID.String())

	resp, err := c.requestDo(req, requestID)

	if err != nil {
		return fmt.Errorf("internal/agent/clients/reporter: send request err: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return ErrUnsuccessfulResponse
	}

	defer resp.Body.Close()

	return nil
}

func (c *Reporter) requestDo(req *http.Request, requestID uuid.UUID) (*http.Response, error) {
	start := time.Now()
	c.logger.Printf("[%s] REQUEST BEGIN", requestID)
	c.logger.Printf("[%s] %s %s", requestID, req.Method, req.URL)
	c.logger.Printf("[%s] Headers: %v", requestID, req.Header)

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Printf("[%s] REQUEST END", requestID)
		return nil, err
	}

	c.logger.Printf("[%s] Duration: %v", requestID, time.Since(start))
	c.logger.Printf("[%s] Response: %s", requestID, resp.Status)
	c.logger.Printf("[%s] REQUEST END", requestID)

	return resp, nil
}

type MetricsBody struct {
	MName string `json:"id"`              // имя метрики
	MType string `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta string `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value string `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m MetricsBody) MarshalJSON() ([]byte, error) {
	type MetricsBodyAlias MetricsBody

	aliasValue := struct {
		MetricsBodyAlias
		Delta int     `json:"delta,omitempty"` // значение метрики в случае передачи counter
		Value float64 `json:"value,omitempty"`
	}{
		MetricsBodyAlias: MetricsBodyAlias(m),
	}

	if m.Value != "" {
		if v, err := strconv.ParseFloat(m.Value, 64); err != nil {
			return nil, err
		} else {
			aliasValue.Value = v
		}
	}

	if m.Delta != "" {
		if v, err := strconv.Atoi(m.Delta); err != nil {
			return nil, err
		} else {
			aliasValue.Delta = v
		}
	}

	return json.Marshal(aliasValue)
}

func prepareBody(mType, mName, value string) (*bytes.Buffer, error) {
	rec := MetricsBody{
		MName: mName,
		MType: mType,
	}

	switch mType {
	case models.GaugeType:
		rec.Value = value
	case models.CounterType:
		rec.Delta = value
	default:
		return nil, fmt.Errorf("internal/agent/clients/reporter.go: underfined type %s", mType)
	}

	buff, err := json.Marshal(rec)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(buff), nil
}
