package server

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/crypto"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func Test_signerMiddleware(t *testing.T) {
	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	if err != nil {
		t.Fatalf("failed to create zap logger: %v", err)
	}

	type fields struct {
		cfg     config.Config
		headers func(*fields) map[string]string
		body    []byte
	}
	tests := []struct {
		name           string
		fields         fields
		request        func(ctx context.Context, f *fields)
		wantStatusCode int
	}{

		{
			name:           "when invalid signature",
			wantStatusCode: http.StatusBadRequest,
			fields: fields{
				cfg: config.Config{
					Key: "secret",
				},
				headers: func(*fields) map[string]string {
					h := map[string]string{}
					h[signHeaderKey] = "invalid"
					return h
				},
			},
		},
		{
			name:           "when valid signature",
			wantStatusCode: http.StatusOK,
			fields: fields{
				cfg: config.Config{
					Key: "secret",
				},
				body: []byte(`{"fiz": "baz"}`),
				headers: func(f *fields) map[string]string {
					secret := f.cfg.Key

					sign, err := crypto.NewCms(hmac.New(sha256.New, []byte(secret))).Sign(bytes.NewBuffer(f.body))
					if err != nil {
						t.Fatalf("create sign error %s", err.Error())
						return nil
					}

					h := map[string]string{}
					h[signHeaderKey] = base64.StdEncoding.EncodeToString(sign)
					return h
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			r.Use(signerMiddleware(lg, []byte(tt.fields.cfg.Key)))
			r.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			srv := httptest.NewServer(r)
			defer srv.Close()

			req, err := http.NewRequest(
				"GET",
				srv.URL,
				bytes.NewBuffer(tt.fields.body),
			)
			assert.NoError(t, err)

			for k, v := range tt.fields.headers(&tt.fields) {
				req.Header.Set(k, v)
			}

			resp, err := srv.Client().Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.NoError(t, resp.Body.Close())
		})
	}
}

func Test_headerMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		headers     map[string]string
		headerAdded bool
	}{
		{
			name:        "when application/json",
			headerAdded: true,
			headers: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name:        "whithout application/json",
			headerAdded: false,
			headers: map[string]string{
				"Content-Type": "plain/text",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			r.Use(headerMiddleware())
			r.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			srv := httptest.NewServer(r)
			defer srv.Close()

			req, err := http.NewRequest(
				"GET",
				srv.URL,
				nil)
			assert.NoError(t, err)

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			resp, err := srv.Client().Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.headerAdded, resp.Header.Get("Content-Type") == "application/json")
			assert.NoError(t, resp.Body.Close())
		})
	}
}

func Test_httpLoggerMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		wantStatusCode int
	}{

		{
			name:           "status code 200",
			wantStatusCode: http.StatusOK,
		},
	}

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	if err != nil {
		t.Fatalf("failed to create zap logger: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			r.Use(httpLoggerMiddleware(lg))
			r.GET("/", func(c *gin.Context) {
				c.Status(tt.wantStatusCode)
			})

			srv := httptest.NewServer(r)
			defer srv.Close()

			req, err := http.NewRequest(
				"GET",
				srv.URL,
				nil)
			assert.NoError(t, err)

			resp, err := srv.Client().Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.NoError(t, resp.Body.Close())
		})
	}
}

func Test_decrypterMiddleware(t *testing.T) {
	type args struct {
		cfg       *config.Config
		decrypter *server.MockDecrypter
	}
	tests := []struct {
		name           string
		args           args
		wantStatusCode int
		prepare        func(*args)
	}{
		{
			name: "when decrypt failed",
			args: args{
				cfg: &config.Config{
					PrivateKey: bytes.NewReader([]byte("secret")),
				},
			},
			prepare: func(args *args) {
				args.decrypter.EXPECT().Decrypt(gomock.Any()).Return("", errors.New("decrypt failed"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "when ok",
			args: args{
				cfg: &config.Config{
					PrivateKey: bytes.NewReader([]byte("secret")),
				},
			},
			prepare: func(args *args) {
				args.decrypter.EXPECT().Decrypt(gomock.Any()).Return("body", nil)
			},
			wantStatusCode: http.StatusOK,
		},
	}

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	if err != nil {
		t.Fatalf("failed to create zap logger: %v", err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			tt.args.decrypter = server.NewMockDecrypter(ctrl)
			tt.prepare(&tt.args)

			r.Use(decrypterMiddleware(lg, tt.args.decrypter))
			r.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			srv := httptest.NewServer(r)
			defer srv.Close()

			req, err := http.NewRequest(
				"GET",
				srv.URL,
				nil)
			assert.NoError(t, err)
			resp, err := srv.Client().Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.NoError(t, resp.Body.Close())
		})
	}
}

func Test_middlewares(t *testing.T) {
	type args struct {
		cfg *config.Config
	}
	tests := []struct {
		name    string
		args    args
		want    []gin.HandlerFunc
		wantErr bool
	}{
		{
			name: "default",
			args: args{cfg: &config.Config{}},
			want: []gin.HandlerFunc{
				gin.Recovery(),
				headerMiddleware(),
				httpLoggerMiddleware(nil),
				gzip.Gzip(gzip.DefaultCompression),
			},
		},
		{
			name: "when private key present",
			args: args{cfg: &config.Config{PrivateKey: &bytes.Buffer{}}},
			want: []gin.HandlerFunc{
				gin.Recovery(),
				headerMiddleware(),
				httpLoggerMiddleware(nil),
				gzip.Gzip(gzip.DefaultCompression),
				decrypterMiddleware(nil, nil),
			},
		},
		{
			name: "when cms key present",
			args: args{cfg: &config.Config{Key: "secret"}},
			want: []gin.HandlerFunc{
				gin.Recovery(),
				headerMiddleware(),
				httpLoggerMiddleware(nil),
				gzip.Gzip(gzip.DefaultCompression),
				signerMiddleware(nil, []byte("")),
			},
		},
		{
			name: "when trusted subnet present",
			args: args{cfg: &config.Config{TrustedSubnet: "192.168.0.1/24"}},
			want: []gin.HandlerFunc{
				gin.Recovery(),
				headerMiddleware(),
				httpLoggerMiddleware(nil),
				gzip.Gzip(gzip.DefaultCompression),
				aclMiddleware(nil, nil),
			},
		},
		{
			name: "when trusted subnet invalid",
			args: args{cfg: &config.Config{TrustedSubnet: "192.168.0/24"}},
			want: []gin.HandlerFunc{
				gin.Recovery(),
				headerMiddleware(),
				httpLoggerMiddleware(nil),
				gzip.Gzip(gzip.DefaultCompression),
				aclMiddleware(nil, nil),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := middlewares(tt.args.cfg, nil, nil)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.want), len(got))
			}
		})
	}
}

func Test_aclMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		ip             string
		wantStatusCode int
	}{
		{
			name:           "when ip is blank",
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:           "when ip invalid",
			ip:             "123",
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:           "when ip does not belond to network",
			ip:             "192.168.1.1",
			wantStatusCode: http.StatusForbidden,
		},
	}

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	if err != nil {
		t.Fatalf("failed to create zap logger: %v", err)
	}

	_, ipnet, err := net.ParseCIDR("192.168.0.1/24")
	assert.NoError(t, err)
	mw := aclMiddleware(lg, ipnet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()

			r.Use(mw)
			r.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			srv := httptest.NewServer(r)
			defer srv.Close()

			req, err := http.NewRequest(
				"GET",
				srv.URL,
				nil)
			req.Header.Add(XRealIPHeader, tt.ip)

			assert.NoError(t, err)
			resp, err := srv.Client().Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.NoError(t, resp.Body.Close())
		})
	}
}
