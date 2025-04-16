package server

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
			name:           "when skip signature",
			wantStatusCode: http.StatusOK,
			fields: fields{
				cfg: config.Config{
					Key: "",
				},
				headers: func(*fields) map[string]string { return map[string]string{} },
			},
		},
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
			r.Use(signerMiddleware(lg, &tt.fields.cfg))
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
			name: "when private key is nil",
			args: args{
				cfg: &config.Config{},
			},
			prepare:        func(args *args) {},
			wantStatusCode: http.StatusOK,
		},
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

			r.Use(decrypterMiddleware(lg, tt.args.cfg, tt.args.decrypter))
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
