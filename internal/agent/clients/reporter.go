package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
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
	req, err := http.NewRequestWithContext(
		context.TODO(),
		"POST", fmt.Sprintf("%s/update/%s/%s/%v", c.address, mType, mName, value),
		http.NoBody,
	)

	if err != nil {
		return fmt.Errorf("internal/agent/clients/reporter: create request err: %w", err)
	}

	req.Header.Add("Content-Type", "text/plain")

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
