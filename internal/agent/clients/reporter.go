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
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent"
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

type IRealIPHeaderSetter interface {
	Call(req *http.Request) error
}

// Reporter handles the communication with the metrics server
type Reporter struct {
	client          Requester
	address         string
	lg              *logging.ZapLogger
	compressor      compressor
	maxAttempts     uint8
	secretKey       []byte
	semaphore       *semaphore
	publicKeyPath   io.Reader
	encryptor       Encryptor
	repository      *agent.MetricsRepository
	ipAddressSetter IRealIPHeaderSetter
}

// NewReporter creates a new Reporter instance with basic configuration
func NewReporter(address string, lg *logging.ZapLogger, client Requester) *Reporter {
	return &Reporter{
		address:     address,
		client:      client,
		lg:          lg,
		maxAttempts: 2,
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

type Encryptor interface {
	Encrypt(b []byte) (string, error)
}

// NewCompReporter creates a new Reporter instance with compression and rate limiting
func NewCompReporter(address string, lg *logging.ZapLogger, cfg *config.Config, client Requester, ips IRealIPHeaderSetter, repository *agent.MetricsRepository) *Reporter {
	reporter := &Reporter{
		address:         address,
		client:          client,
		lg:              lg,
		compressor:      gzbody,
		maxAttempts:     cfg.MaxAttempts,
		secretKey:       []byte(cfg.Key),
		semaphore:       NewSemaphore(cfg.RateLimit),
		publicKeyPath:   cfg.HTTPCert,
		ipAddressSetter: ips,
		repository:      repository,
	}

	if cfg.HTTPCert != nil {
		reporter.encryptor = crypto.NewEncryptor(cfg.HTTPCert)
	}

	return reporter
}

// UpdateMetric sends a single metric update to the server
func (c *Reporter) UpdateMetric(ctx context.Context, mType, mName, value string) error {
	reqCtx := c.lg.WithContextFields(ctx, zap.String("name", "http"))
	reqCtx = context.WithoutCancel(reqCtx)
	body, err := c.prepareBody(mType, mName, value)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST", fmt.Sprintf("%s/update/", c.address),
		body,
	)

	if err != nil {
		return fmt.Errorf("internal/agent/clients/reporter: create request err: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Request-ID", uuid.NewV4().String())

	resp, err := c.processRequest(reqCtx, req)
	defer func() {
		if resp != nil {
			if closeErr := resp.Body.Close(); err != nil {
				c.lg.ErrorCtx(reqCtx, "close body erorr", zap.Error(closeErr))
			}
		}
	}()

	if err != nil {
		return fmt.Errorf("internal/agent/clients/reporter: send request err: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return ErrUnsuccessfulResponse
	}

	return nil
}

// UpdateMetrics sends a batch of metric updates to the server
func (c *Reporter) UpdateMetrics(ctx context.Context, data []*models.Metric) error {
	reqCtx := c.lg.WithContextFields(ctx, zap.String("name", "http"))
	reqCtx = context.WithoutCancel(reqCtx)

	metricsBody := make([]MetricsBody, 0, len(data))
	for _, m := range data {
		name, mType, value := c.repository.SafeRead(m)
		rec, err := generateMetric(mType, name, value)
		if err != nil {
			return fmt.Errorf("internal/agent/clients/reporter: generate metric %+v error %w", m, err)
		}

		metricsBody = append(metricsBody, rec)
	}

	var body bytes.Buffer

	if err := json.NewEncoder(&body).Encode(metricsBody); err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST", fmt.Sprintf("%s/updates/", c.address),
		&body,
	)
	if err != nil {
		return fmt.Errorf("internal/agent/clients/reporter: create request err: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Request-ID", uuid.NewV4().String())
	resp, err := c.processRequest(reqCtx, req)
	defer func() {
		if resp != nil {
			if closeErr := resp.Body.Close(); err != nil {
				c.lg.ErrorCtx(reqCtx, "close body erorr", zap.Error(closeErr))
			}
		}
	}()

	if err != nil {
		return fmt.Errorf("internal/agent/clients/reporter: send request err: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return ErrUnsuccessfulResponse
	}

	return nil
}

func generateMetric(mType, mName, mValue string) (MetricsBody, error) {
	rec := MetricsBody{MName: mName, MType: mType}
	switch mType {
	case models.GaugeType:
		rec.Value = mValue
	case models.CounterType:
		rec.Delta = mValue
	default:
		return MetricsBody{}, fmt.Errorf("internal/agent/clients/reporter.go: underfined type: %s, for name: %s, value: %s", mType, mName, mValue)
	}

	return rec, nil
}

type bodyKey string

var bKey bodyKey = "body"

const XRealIPHeader = "X-Real-IP"

func (c *Reporter) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	request := func() (*http.Response, error) {
		c.semaphore.Acquire()
		defer c.semaphore.Release()

		resp, err := c.client.Request(req)
		if err != nil {
			return nil, fmt.Errorf("reporter: request failed error %w", err)
		}

		if err := resp.Body.Close(); err != nil {
			return nil, fmt.Errorf("reporter: response body close error %w", err)
		}

		return resp, nil
	}

	var (
		resp *http.Response
		err  error
	)
	for i := range c.maxAttempts {
		c.lg.DebugCtx(ctx, "request", zap.Uint8("attempt", i+1), zap.Uint16("limit", uint16(c.maxAttempts)))
		resp, err = request()
		if err == nil {
			return resp, nil
		}

		if i == c.maxAttempts-1 {
			break
		}

		del := attemptDelay(i)
		runNext := time.Now().Add(del).Format(time.RFC3339Nano)
		c.lg.DebugCtx(
			ctx,
			"request failed",
			zap.Uint8("attempt", i+1),
			zap.Uint16("limit", uint16(c.maxAttempts)),
			zap.String("run next", runNext),
		)
		time.Sleep(del)
	}

	return nil, fmt.Errorf("reporter: request max attempts exceeded errpr %w", err)
}

func (c *Reporter) processRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	buff := bytes.Buffer{}

	reader, readerErr := req.GetBody()
	if readerErr != nil {
		return nil, fmt.Errorf("internal/agent/clients/reporter read body error %w", readerErr)
	}

	if _, copyErr := io.Copy(&buff, reader); copyErr != nil {
		return nil, fmt.Errorf("internal/agent/clients/reporter read body error %w", copyErr)
	}

	reqCtx := context.WithValue(ctx, bKey, buff.Bytes())

	if signErr := c.signRequest(reqCtx, req); signErr != nil {
		return nil, fmt.Errorf("internal/agent/clients/reporter sign request error %w", signErr)
	}

	if encErr := c.encryptRequest(reqCtx, req); encErr != nil {
		return nil, fmt.Errorf("internal/agent/clients/reporter encrypt request error %w", encErr)
	}

	if setIpErr := c.ipAddressSetter.Call(req); setIpErr != nil {
		return nil, fmt.Errorf("internal/agent/clients/reporter set ip address error %w", setIpErr)
	}

	resp, reqErr := c.doRequest(reqCtx, req)
	if reqErr != nil {
		return nil, fmt.Errorf("internal/agent/clients/reporter send request error %w", reqErr)
	}

	return resp, nil
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

func (c *Reporter) encryptRequest(ctx context.Context, r *http.Request) error {
	if c.encryptor == nil {
		return nil
	}

	b, ok := ctx.Value(bKey).([]byte)
	if !ok {
		return ErrBodyKeyNotFoundInContext
	}

	encrypted, err := c.encryptor.Encrypt(b)
	if err != nil {
		return fmt.Errorf("internal/agent/clients/reporter encrypt request error %w", err)
	}

	r.Body = io.NopCloser(bytes.NewBuffer([]byte(encrypted)))

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
	aliasValue := MetricsBodyAlias{
		MetricsBody: m,
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
	res := &bytes.Buffer{}
	w, err := flate.NewWriter(res, flate.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("internal/agent/clients/reporter.go write to buffer error %w", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("internal/agent/clients/reporter.go close writer error %w", err)
	}

	return res, nil
}

func attemptDelay(i uint8) time.Duration {
	delay := int(utils.Delay(i) * 1000)

	return time.Duration(delay) * time.Millisecond
}
