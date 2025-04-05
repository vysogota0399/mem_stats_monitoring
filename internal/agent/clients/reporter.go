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

	"github.com/mailru/easyjson"
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

// Requester interface defines the contract for making HTTP requests
type Requester interface {
	Request(r *http.Request) (*http.Response, error)
}

// Reporter handles the communication with the metrics server
type Reporter struct {
	client      Requester
	address     string
	lg          *logging.ZapLogger
	compressor  compressor
	maxAttempts uint8
	secretKey   []byte
	semaphore   *semaphore
}

// NewReporter creates a new Reporter instance with basic configuration
func NewReporter(address string, lg *logging.ZapLogger, client Requester) *Reporter {
	return &Reporter{
		address:     address,
		client:      client,
		lg:          lg,
		maxAttempts: 5,
	}
}

// semaphore provides a simple semaphore implementation for rate limiting
type semaphore struct {
	semaCh chan struct{}
}

// Acquire acquires a semaphore permit
func (s *semaphore) Acquire() {
	s.semaCh <- struct{}{}
}

// Release releases a semaphore permit
func (s *semaphore) Release() {
	<-s.semaCh
}

// NewSemaphore creates a new semaphore with the specified maximum number of permits
func NewSemaphore(maxReq int) *semaphore {
	return &semaphore{
		semaCh: make(chan struct{}, maxReq),
	}
}

// NewCompReporter creates a new Reporter instance with compression and rate limiting
func NewCompReporter(address string, lg *logging.ZapLogger, cfg *config.Config, client Requester) *Reporter {
	return &Reporter{
		address:     address,
		client:      client,
		lg:          lg,
		compressor:  gzbody,
		maxAttempts: cfg.MaxAttempts,
		secretKey:   []byte(cfg.Key),
		semaphore:   NewSemaphore(cfg.RateLimit),
	}
}

// UpdateMetric sends a single metric update to the server
func (c *Reporter) UpdateMetric(ctx context.Context, mType, mName, value string) error {
	ctx = c.lg.WithContextFields(ctx, zap.String("name", "http"))
	body, err := c.prepareBody(mType, mName, value)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.lg.ErrorCtx(ctx, "failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return ErrUnsuccessfulResponse
	}

	return nil
}

// UpdateMetrics sends a batch of metric updates to the server
func (c *Reporter) UpdateMetrics(ctx context.Context, data []*models.Metric) error {
	metricsBody := make([]MetricsBody, 0, len(data))
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
		ctx,
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.lg.ErrorCtx(ctx, "failed to close response body", zap.Error(closeErr))
		}
	}()

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
		return nil, fmt.Errorf("internal/agent/clients/reporter.go: underfined type: %s, for name: %s, value: %s", mType, mName, mValue)
	}

	return rec, nil
}

type bodyKey string

var bKey bodyKey = "body"

func (c *Reporter) requestDo(ctx context.Context, req *http.Request) (*http.Response, error) {
	resCh := make(chan *http.Response, c.maxAttempts)
	errCh := make(chan error)
	defer close(resCh)
	defer close(errCh)

	doRequest := func(ctx context.Context, req *http.Request, resChan chan *http.Response, errChan chan error) {
		c.semaphore.Acquire()
		defer c.semaphore.Release()

		start := time.Now()
		reader, readerErr := req.GetBody()
		if readerErr != nil {
			errChan <- fmt.Errorf("internal/agent/clients/reporter get body reader error %w", readerErr)
			return
		}

		buff := bytes.Buffer{}
		if _, copyErr := io.Copy(&buff, reader); copyErr != nil {
			errChan <- fmt.Errorf("internal/agent/clients/reporter read body error %w", copyErr)
			return
		}

		ctx = context.WithValue(ctx, bKey, buff.Bytes())

		if signErr := c.signRequest(ctx, req); signErr != nil {
			errChan <- fmt.Errorf("internal/agent/clients/reporter sign request error %w", signErr)
			return
		}

		resp, reqErr := c.client.Request(req)
		if reqErr != nil {
			errChan <- fmt.Errorf("internal/agent/clients/reporter send request error %w", reqErr)
			return
		}
		defer resp.Body.Close()

		c.lg.DebugCtx(ctx,
			"request",
			zap.String("method", req.Method),
			zap.String("url", req.URL.String()),
			zap.String("body", buff.String()),
			zap.Duration("duration", time.Since(start)),
			zap.String("status", resp.Status),
			zap.Any("headers", req.Header),
		)

		resChan <- resp
	}

	var err error

	for i := uint8(0); i < c.maxAttempts; i++ {
		go doRequest(ctx, req, resCh, errCh)

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("internal/agent/clients/reporter context canceled")
		case res := <-resCh:
			return res, nil
		case err = <-errCh:
			if i >= c.maxAttempts-1 {
				return nil, err
			}

			c.lg.DebugCtx(ctx, "wait next", zap.Int("total", int(c.maxAttempts)), zap.Int("current", int(i+1)))
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

// MetricsBody represents the structure of a metric in the request body
type MetricsBody struct {
	MName string `json:"id"`              // metric name
	MType string `json:"type"`            // metric type (gauge or counter)
	Delta string `json:"delta,omitempty"` // metric value for counter type
	Value string `json:"value,omitempty"` // metric value for gauge type
}

// MetricsBodyAlias provides type conversion for metric values
type MetricsBodyAlias struct {
	MetricsBody
	Delta int     `json:"delta,omitempty"` // metric value for counter type (converted to int)
	Value float64 `json:"value,omitempty"` // metric value for gauge type (converted to float64)
}

func (m MetricsBody) MarshalJSON() ([]byte, error) {
	aliasValue := struct {
		MetricsBodyAlias
		Delta int     `json:"delta,omitempty"` // metric value for counter type
		Value float64 `json:"value,omitempty"` // metric value for gauge type
	}{
		MetricsBodyAlias: MetricsBodyAlias{
			MetricsBody: m,
		},
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

	return easyjson.Marshal(aliasValue)
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
