// Модуль server отвечает за инициализацию и запуск web - сервера. В нем определены эндпоинты, обработчики и middleware.
package server

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/crypto"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

func TestNewServer(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{}
	strg := storage.NewMemory()
	lg, err := logging.MustZapLogger(zapcore.Level(cfg.LogLevel))
	assert.NoError(t, err)
	srvc := service.New(strg, lg)

	srv := NewServer(ctx, cfg, strg, srvc, lg, withTestServer)
	assert.Equal(t, cfg, srv.config)
	assert.Equal(t, strg, srv.storage)
	assert.Equal(t, srvc, srv.service)
	assert.Equal(t, lg, srv.lg)
}

func TestServer_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Config{Address: "localhost:8080"}
	strg := storage.NewMemory()
	lg, err := logging.MustZapLogger(zapcore.Level(cfg.LogLevel))
	assert.NoError(t, err)

	srvc := service.New(strg, lg)

	errg, ctx := errgroup.WithContext(ctx)
	srv := NewServer(ctx, cfg, strg, srvc, lg, withTestServer)
	srv.htmlRoute = false
	srv.Start(errg)

	cancel()
	assert.NoError(t, errg.Wait())
}

func TestNewTestServer(t *testing.T) {
	type args struct {
		cfg config.Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{cfg: config.Config{Address: "localhost:3000"}},
			name: "when valid config",
		},
		{
			args:    args{cfg: config.Config{}},
			name:    "when invalid config",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTestServer(tt.args.cfg)
			assert.NoError(t, err)
		})
	}
}

func TestServer_signer(t *testing.T) {
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
					Address: ":1234",
					Key:     "",
				},
				headers: func(*fields) map[string]string { return map[string]string{} },
			},
		},
		{
			name:           "when invalid signature",
			wantStatusCode: http.StatusBadRequest,
			fields: fields{
				cfg: config.Config{
					Address: ":1234",
					Key:     "secret",
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
					Address: ":1234",
					Key:     "secret",
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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			strg := storage.NewMemory()
			lg, err := logging.MustZapLogger(zapcore.Level(tt.fields.cfg.LogLevel))
			assert.NoError(t, err)
			srvc := service.New(strg, lg)
			errg, ctx := errgroup.WithContext(ctx)

			srv := NewServer(ctx, tt.fields.cfg, strg, srvc, lg, withTestServer)
			srv.htmlRoute = false
			srv.router.GET("/test", func(ctx *gin.Context) {})
			srv.Start(errg)

			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodGet,
				fmt.Sprintf("http://localhost%s/test", tt.fields.cfg.Address),
				bytes.NewBuffer(tt.fields.body),
			)
			assert.NoError(t, err)

			req.Header.Add("Content-Type", "application/json")

			for k, v := range tt.fields.headers(&tt.fields) {
				req.Header.Add(k, v)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}

func Test_headers(t *testing.T) {
	tests := []struct {
		name           string
		headerPresence bool
		headers        map[string]string
	}{
		{
			name:           "when content type defined",
			headerPresence: true,
			headers:        map[string]string{"Content-Type": "application/json"},
		},
		{
			name:           "when content type not defined",
			headerPresence: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			strg := storage.NewMemory()
			lg, err := logging.MustZapLogger(zapcore.Level(zap.DebugLevel))
			assert.NoError(t, err)
			srvc := service.New(strg, lg)
			errg, ctx := errgroup.WithContext(ctx)

			cfg := config.Config{Address: ":1234"}
			srv := NewServer(ctx, cfg, strg, srvc, lg, withTestServer)
			srv.htmlRoute = false
			srv.router.GET("/test", func(ctx *gin.Context) {})
			srv.Start(errg)

			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodGet,
				fmt.Sprintf("http://localhost%s/test", cfg.Address),
				&bytes.Buffer{},
			)
			assert.NoError(t, err)
			for k, v := range tt.headers {
				req.Header.Add(k, v)
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			if tt.headerPresence {
				assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
			} else {
				assert.Equal(t, "", resp.Header.Get("Content-Type"))
			}
		})
	}
}
