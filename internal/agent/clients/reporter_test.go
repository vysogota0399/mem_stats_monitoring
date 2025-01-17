package clients

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *Reporter {
	testClient := &http.Client{
		Transport: fn,
	}

	lg, _ := logging.MustZapLogger(zapcore.DebugLevel)
	return &Reporter{
		address: "http://0.0.0.0:8080",
		client:  testClient,
		lg:      lg,
	}
}

func TestNewReporter(t *testing.T) {
	tasks := []struct {
		name       string
		err        error
		ftransport RoundTripFunc
	}{
		{
			name: "when 200",
			err:  nil,
			ftransport: func(r *http.Request) *http.Response {
				assert.Equal(t, r.Header.Get("Content-Type"), "application/json")
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
				assert.Equal(t, r.Header.Get("Content-Type"), "application/json")
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

			assert.Equal(t, tt.err, client.UpdateMetric(context.TODO(), models.CounterType, "baz", "1"))
		})
	}
}

func Test_prepareBody(t *testing.T) {
	type args struct {
		mType string
		mName string
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    *bytes.Buffer
		wantErr bool
	}{
		{
			args: args{
				mType: models.CounterType,
				mName: "test",
				value: "1",
			},
			name:    "when counter",
			wantErr: false,
		},
		{
			args: args{
				mType: models.GaugeType,
				mName: "test",
				value: "1",
			},
			name:    "when gauge",
			wantErr: false,
		},
		{
			args: args{
				mType: "underfined",
				mName: "",
				value: "",
			},
			name:    "when underfined",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewTestClient(func(req *http.Request) *http.Response { return nil })
			_, err := client.prepareBody(tt.args.mType, tt.args.mName, tt.args.value)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
