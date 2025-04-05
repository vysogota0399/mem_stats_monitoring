// Package server handles the initialization and operation of the web server.
// It defines endpoints, handlers, and middleware for the metrics collection service.
package server

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/handlers"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/crypto"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// IServer defines the interface for server operations
type IServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

// Server represents the main server instance that handles HTTP requests
type Server struct {
	config    config.Config
	router    *gin.Engine
	storage   storage.Storage
	service   *service.Service
	ctx       context.Context
	lg        *logging.ZapLogger
	secretKey []byte
	htmlRoute bool
	server    IServer
}

// NewServer creates a new Server instance with the specified configuration
func NewServer(
	ctx context.Context,
	c config.Config,
	strg storage.Storage,
	srvc *service.Service,
	lg *logging.ZapLogger,
	opts ...ServerOption,
) *Server {
	s := Server{
		config:    c,
		router:    gin.New(),
		service:   srvc,
		ctx:       lg.WithContextFields(ctx, zap.String("name", "server")),
		lg:        lg,
		secretKey: []byte(c.Key),
		htmlRoute: true,
	}

	s.router.Use(
		gin.Recovery(),
		headers(),
		httpLogger(ctx, lg),
		s.signer(),
		gzip.Gzip(gzip.DefaultCompression),
	)

	s.server = &http.Server{
		Addr:              c.Address,
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           s.router,
	}

	s.storage = strg
	return &s
}

// Start launches the server in a separate goroutine with graceful shutdown support
func (s *Server) Start(wg *errgroup.Group) {
	pprof.Register(s.router)
	if s.htmlRoute {
		s.router.LoadHTMLGlob("internal/server/templates/*.tmpl")
		s.router.GET("/", handlers.NewRootHandler(s.storage))
	}

	s.router.POST("/update/:type/:name/:value", handlers.NewUpdateMetricHandler(s.storage))
	s.router.POST("/update/", handlers.NewRestUpdateMetricHandler(s.storage, s.service.UpdateMetricService, s.lg))
	s.router.POST("/updates/", handlers.NewUpdatesRestMetricsHandler(s.storage, s.service.UpdateMetricsService, s.lg))
	s.router.POST("/value/", handlers.NewShowRestMetricHandler(s.storage, s.lg))
	s.router.GET("/value/:type/:name", handlers.NewShowMetricHandler(s.storage))
	s.router.GET("/ping", handlers.NewPingHandler(s.storage, s.lg))

	s.lg.DebugCtx(s.ctx, "start", zap.String("config", s.config.String()))

	wg.Go(func() error {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			os.Exit(-1)
			return fmt.Errorf("server: startup server failed error %w", err)
		}

		return nil
	})

	wg.Go(func() error {
		<-s.ctx.Done()

		s.lg.DebugCtx(s.ctx, "Graceful shutdown")
		if err := s.server.Shutdown(s.ctx); err != nil {
			return fmt.Errorf("server: gracefull shutdown failed error %w", err)
		}

		return nil
	})
}

// headers middleware sets appropriate content type headers for JSON responses
func headers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("Content-Type") == "application/json" {
			c.Writer.Header().Set("Content-Type", "application/json")
		}

		c.Next()
	}
}

// httpLogger middleware logs request/response details including timing and body content
func httpLogger(ctx context.Context, lg *logging.ZapLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		var requestID string
		if reqID := c.Request.Header.Get("X-Request-ID"); reqID != "" {
			requestID = reqID
		} else {
			requestID = uuid.NewV4().String()
		}

		ctx = lg.WithContextFields(ctx,
			zap.String("request_id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.Reflect("params", c.Params),
			zap.Reflect("headers", c.Request.Header),
		)

		bodyBuff := &bytes.Buffer{}
		if _, err := io.Copy(bodyBuff, c.Request.Body); err != nil {
			lg.ErrorCtx(ctx, "read body error", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Body = io.NopCloser(bodyBuff)
		c.Set("request_id", requestID)

		lg.InfoCtx(ctx, "request", zap.String("body", bodyBuff.String()))

		c.Next()

		status := c.Writer.Status()
		bodySize := c.Writer.Size()
		lg.InfoCtx(ctx, "response",
			zap.String("body", bodyBuff.String()),
			zap.Int("status", status),
			zap.Int("response_size", bodySize),
			zap.Duration("duration", time.Since(start)),
		)
	}
}

var signHeaderKey string = "HashSHA256"

// signResponseReadWriter wraps gin.ResponseWriter to capture response body for signing
type signResponseReadWriter struct {
	gin.ResponseWriter
	b *bytes.Buffer
}

// Write implements the io.Writer interface
func (rw *signResponseReadWriter) Write(b []byte) (int, error) {
	rw.b.Write(b)
	return rw.ResponseWriter.Write(b)
}

// Read implements the io.Reader interface
func (rw *signResponseReadWriter) Read(b []byte) (int, error) {
	return rw.b.Read(b)
}

// signer middleware verifies request signatures and signs responses
func (s *Server) signer() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(s.secretKey) == 0 {
			c.Next()
			return
		}

		rw := &signResponseReadWriter{b: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = rw

		cms := crypto.NewCms(hmac.New(sha256.New, s.secretKey))
		base64sign := c.Request.Header.Get(signHeaderKey)
		if base64sign == "" {
			c.Next()
			return
		}

		sign, err := base64.StdEncoding.DecodeString(base64sign)
		if err != nil {
			s.lg.ErrorCtx(c, "base64 decode sign error", zap.Error(err))
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		bodyBuff := &bytes.Buffer{}
		if _, err := io.Copy(bodyBuff, c.Request.Body); err != nil {
			s.lg.ErrorCtx(c, "read body error", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Body = io.NopCloser(bodyBuff)

		if eq, err := cms.Verify(bodyBuff, sign); err != nil || !eq {
			s.lg.ErrorCtx(c, "invalid request signature", zap.Error(err))
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.Next()

		sign, err = cms.Sign(rw)
		if err != nil {
			s.lg.ErrorCtx(c, "response sign failed", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Header.Set(signHeaderKey, base64.StdEncoding.EncodeToString(sign))
	}
}

// ServerOption defines a function type for configuring Server instances
type ServerOption func(s *Server) error

// TestServer provides a test implementation of the IServer interface
type TestServer struct {
	*httptest.Server
}

// NewTestServer creates a new TestServer instance for testing purposes
func NewTestServer(cfg config.Config) (*TestServer, error) {
	ls, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("server: create test listener error %w", err)
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewUnstartedServer(h)
	srv.Listener = ls

	return &TestServer{Server: srv}, nil
}

// ListenAndServe starts the test server
func (srv *TestServer) ListenAndServe() error {
	srv.Start()
	return nil
}

// Shutdown gracefully stops the test server
func (srv *TestServer) Shutdown(ctx context.Context) error {
	srv.Close()
	return nil
}

var withTestServer ServerOption = func(s *Server) error {
	srv, err := NewTestServer(s.config)
	if err != nil {
		return fmt.Errorf("server: create test server error %w", err)
	}

	s.server = srv
	return nil
}
