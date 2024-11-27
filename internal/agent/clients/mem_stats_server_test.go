package clients

import (
	"net/http"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *MemStatsServer {
	testClient := &http.Client{
		Transport: RoundTripFunc(fn),
	}

	return &MemStatsServer{
		address: "http://0.0.0.0:8080",
		client:  testClient,
		logger:  utils.InitLogger("[test]"),
	}
}

func TestNewMemStatsServer(t *testing.T) {
	tasks := []struct {
		name       string
		err        error
		ftransport RoundTripFunc
	}{
		{
			name: "when 200",
			err:  nil,
			ftransport: func(r *http.Request) *http.Response {
				assert.Equal(t, r.Header.Get("Content-Type"), "text/plain")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
				}
			},
		},
		{
			name: "when not 200",
			err:  ErrUnsuccessfulResponse,
			ftransport: func(r *http.Request) *http.Response {
				assert.Equal(t, r.Header.Get("Content-Type"), "text/plain")
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Header:     make(http.Header),
				}
			},
		},
	}

	for _, tt := range tasks {
		t.Run(tt.name, func(t *testing.T) {
			client := NewTestClient(tt.ftransport)

			assert.Equal(t, tt.err, client.UpdateMetric("fiz", "baz", "1", uuid.NewV4()))
		})
	}
}
