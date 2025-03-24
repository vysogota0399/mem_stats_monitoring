package clients

import (
	"bytes"
	"compress/flate"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/crypto"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

var ErrUnsuccessfulResponse error = errors.New("mem_stats_server: response not success")

type compressor func(*bytes.Buffer) (*bytes.Buffer, error)

type Requester interface {
	Request(r *http.Request) (*http.Response, error)
}

type Reporter struct {
	client     Requester
	address    string
	lg         *logging.ZapLogger
	compressor compressor
	maxRetries uint8
	secretKey  []byte
	semaphore  *semaphore
}

func NewReporter(address string, lg *logging.ZapLogger, client Requester) *Reporter {
	return &Reporter{
		address:    address,
		client:     client,
		lg:         lg,
		maxRetries: 5,
	}
}

type semaphore struct {
	semaCh chan struct{}
}

func (s *semaphore) Acquire() {
	s.semaCh <- struct{}{}
}

func (s *semaphore) Release() {
	<-s.semaCh
}

func NewSemaphore(maxReq int) *semaphore {
	return &semaphore{
		semaCh: make(chan struct{}, maxReq),
	}
}

func NewCompReporter(address string, lg *logging.ZapLogger, cfg *config.Config, client Requester) *Reporter {
	return &Reporter{
		address:    address,
		client:     client,
		lg:         lg,
		compressor: gzbody,
		maxRetries: 5,
		secretKey:  []byte(cfg.Key),
		semaphore:  NewSemaphore(cfg.RateLimit),
	}
}

func (c *Reporter) UpdateMetric(ctx context.Context, mType, mName, value string) error {
	ctx = c.lg.WithContextFields(ctx, zap.String("name", "http"))
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

func (c *Reporter) UpdateMetrics(ctx context.Context, data []*models.Metric) error {
	var metricsBody []MetricsBody
	for _, m := range data {
		rec, err := generateMetric(m.Type, m.Name, m.Value)
		if err != nil {
			return err
		}

		metricsBody = append(metricsBody, *rec)
	}

	var body bytes.Buffer

	if err := json.NewEncoder(&body).Encode(metricsBody); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		context.TODO(),
		"POST", fmt.Sprintf("%s/updates/", c.address),
		&body,
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

func generateMetric(mType, mName, mValue string) (*MetricsBody, error) {
	rec := &MetricsBody{MName: mName, MType: mType}
	switch mType {
	case models.GaugeType:
		rec.Value = mValue
	case models.CounterType:
		rec.Delta = mValue
	default:
		return nil, fmt.Errorf("internal/agent/clients/reporter.go: underfined type %s", mType)
	}

	return rec, nil
}

type bodyKey string

var bKey bodyKey = "body"

func (c *Reporter) requestDo(ctx context.Context, req *http.Request) (*http.Response, error) {
	resCh := make(chan *http.Response, c.maxRetries)
	errCh := make(chan error)
	doRequest := func(ctx context.Context, req *http.Request, resChan chan *http.Response, errChan chan error) {
		c.semaphore.Acquire()
		defer c.semaphore.Release()

		start := time.Now()
		reader, err := req.GetBody()
		if err != nil {
			errChan <- fmt.Errorf("internal/agent/clients/reporter get body reader error %w", err)
			return
		}

		b, err := io.ReadAll(reader)
		if err != nil {
			errChan <- fmt.Errorf("internal/agent/clients/reporter read body error %w", err)
			return
		}

		ctx = context.WithValue(ctx, bKey, b)

		if err := c.signRequest(ctx, req); err != nil {
			errChan <- fmt.Errorf("internal/agent/clients/reporter sign request error %w", err)
			return
		}

		resp, err := c.client.Request(req)
		if err != nil {
			errChan <- fmt.Errorf("internal/agent/clients/reporter send request error %w", err)
			return
		}
		defer resp.Body.Close()

		c.lg.DebugCtx(ctx,
			"request",
			zap.String("method", req.Method),
			zap.String("url", req.URL.String()),
			zap.String("body", string(b)),
			zap.Duration("duration", time.Since(start)),
			zap.String("status", resp.Status),
			zap.Any("headers", req.Header),
		)

		resChan <- resp
	}

	var err error

	go doRequest(ctx, req, resCh, errCh)

	for i := 0; i < int(c.maxRetries); i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("internal/agent/clients/reporter context canceled")
		case res := <-resCh:
			return res, nil
		case err = <-errCh:
			c.lg.DebugCtx(ctx, "wait next", zap.Int("total", int(c.maxRetries)), zap.Int("current", i+1))
			time.Sleep(time.Duration(utils.Delay(uint8(i))) * time.Second)
			go doRequest(ctx, req, resCh, errCh)
		}
	}

	return nil, err
}

var ErrBodyKeyNotFoundInContext error = fmt.Errorf("internal/agent/clients/reporter.go there is no key %s in current context", bKey)
var signHeaderKey string = "HashSHA256"

func (c *Reporter) signRequest(ctx context.Context, r *http.Request) error {
	if len(c.secretKey) == 0 {
		return nil
	}

	b, ok := ctx.Value(bKey).([]byte)
	if !ok {
		return ErrBodyKeyNotFoundInContext
	}

	sign, err := crypto.NewCms(hmac.New(sha256.New, c.secretKey)).Sign(bytes.NewReader(b))
	if err != nil {
		return err
	}

	r.Header.Add(signHeaderKey, base64.StdEncoding.EncodeToString(sign))

	return nil
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
	rec, err := generateMetric(mType, mName, value)
	if err != nil {
		return nil, err
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
