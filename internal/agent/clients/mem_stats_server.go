package clients

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

var ErrUnsuccessfulResponse error = errors.New("mem_stats_server: response not success")

type MemStatsServer struct {
	client  *http.Client
	address string
	logger  utils.Logger
}

func NewMemStatsServer(address string) *MemStatsServer {
	return &MemStatsServer{address: address, client: &http.Client{}, logger: utils.InitLogger("[http]")}
}

func (c *MemStatsServer) UpdateMetric(mType, mName, value string, requestID uuid.UUID) error {
	req, err := http.NewRequest(
		"POST", fmt.Sprintf("%s/update/%s/%s/%v", c.address, mType, mName, value), nil,
	)

	if err != nil {
		c.logger.Printf("[%v] mem_stats_server.go: %v", requestID, err)
		return err
	}

	req.Header.Add("Content-Type", "text/plain")

	resp, err := c.requestDo(req, requestID)

	if err != nil {
		c.logger.Printf("[%v] mem_stats_server: do request error: %v", requestID, err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return ErrUnsuccessfulResponse
	}

	defer resp.Body.Close()

	return nil
}

func (c *MemStatsServer) requestDo(req *http.Request, requestID uuid.UUID) (*http.Response, error) {
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
