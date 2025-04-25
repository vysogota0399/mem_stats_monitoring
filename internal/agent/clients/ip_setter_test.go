package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIpSetter_Call(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "set ip to header",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			srv := httptest.NewServer(h)
			defer srv.Close()

			ips := NewIpSetter(nil)
			r, err := http.NewRequestWithContext(context.Background(), "GET", srv.URL, nil)
			assert.NoError(t, err)
			err = ips.Call(r)
			assert.NoError(t, err)
			assert.NotNil(t, r.Header.Get(XRealIPHeader))
		})
	}
}
