package clients

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

var ErrUnsuccessfulResponse error = errors.New("mem_stats_server: response not success")

type compressor func(*bytes.Buffer) (*bytes.Buffer, error)
type Reporter struct {
	ctx        context.Context
	client     *http.Client
	address    string
	lg         *logging.ZapLogger
	compressor compressor
}

func NewReporter(ctx context.Context, address string, lg *logging.ZapLogger) *Reporter {
	return &Reporter{
		address: address,
		client:  &http.Client{},
		lg:      lg,
		ctx:     lg.WithContextFields(ctx, zap.String("name", "http")),
	}
}

func NewCompReporter(ctx context.Context, address string, lg *logging.ZapLogger) *Reporter {
	return &Reporter{
		address:    address,
		client:     &http.Client{},
		lg:         lg,
		ctx:        lg.WithContextFields(ctx, zap.String("name", "http")),
		compressor: gzbody,
	}
}

func (c *Reporter) UpdateMetric(ctx context.Context, mType, mName, value string) error {
	body, err := c.prepareBody(mType, mName, value)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		context.TODO(),
		"POST", fmt.Sprintf("%s/update/", c.address),
		body,
	)

	if err != nil {
		return fmt.Errorf("internal/agent/clients/reporter: create request err: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Request-ID", uuid.NewV4().String())

	resp, err := c.requestDo(ctx, req)
	if err != nil {
		return fmt.Errorf("internal/agent/clients/reporter: send request err: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrUnsuccessfulResponse
	}

	return nil
}

func (c *Reporter) requestDo(ctx context.Context, req *http.Request) (*http.Response, error) {
	start := time.Now()

	reader, err := req.GetBody()
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	c.lg.DebugCtx(ctx,
		"request",
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
		zap.String("body", string(b)),
		zap.Duration("duration", time.Since(start)),
		zap.String("status", resp.Status),
		zap.Any("headers", req.Header),
	)

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
		Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
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

func (c Reporter) prepareBody(mType, mName, value string) (*bytes.Buffer, error) {
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

func gzbody(b *bytes.Buffer) (*bytes.Buffer, error) {
	var res *bytes.Buffer
	w, err := flate.NewWriter(b, flate.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(res.Bytes())
	if err != nil {
		return nil, fmt.Errorf("internal/agent/clients/reporter.go write to buffer error %w", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("internal/agent/clients/reporter.go close writer error %w", err)
	}

	return res, nil
}
