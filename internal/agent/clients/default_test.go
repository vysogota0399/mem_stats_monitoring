package clients

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaulut(t *testing.T) {
	client := NewDefaulut()
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
}

func TestDefault_Request(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Create test cases
	tests := []struct {
		name           string
		request        *http.Request
		expectedStatus int
		expectedBody   string
		wantErr        bool
	}{
		{
			name:           "successful request",
			request:        &http.Request{Method: "GET", URL: &url.URL{Scheme: "http", Host: server.Listener.Addr().String()}},
			expectedStatus: http.StatusOK,
			expectedBody:   "test response",
			wantErr:        false,
		},
		{
			name:           "invalid request",
			request:        &http.Request{Method: "GET", URL: &url.URL{Scheme: "http", Host: "invalid-url"}},
			expectedStatus: 0,
			expectedBody:   "",
			wantErr:        true,
		},
	}

	client := NewDefaulut()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Request(tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			defer resp.Body.Close()

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body := make([]byte, len(tt.expectedBody))
			resp.Body.Read(body)
			assert.Equal(t, tt.expectedBody, string(body))
		})
	}
}
