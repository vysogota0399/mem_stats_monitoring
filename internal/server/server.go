package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/handlers"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type Server struct {
	config  config.Config
	router  *gin.Engine
	storage storage.Storage
	service *service.Service
	ctx     context.Context
	lg      *logging.ZapLogger
}

type NewServerOption func(*Server)

func NewServer(ctx context.Context, c config.Config, strg storage.Storage, srvc *service.Service, lg *logging.ZapLogger) (*Server, error) {
	s := Server{
		config:  c,
		router:  gin.New(),
		service: srvc,
		ctx:     lg.WithContextFields(ctx, zap.String("name", "server")),
		lg:      lg,
	}

	s.router.Use(
		gin.Recovery(),
		httpLogger(ctx, lg),
		gzip.Gzip(gzip.DefaultCompression),
	)

	s.storage = strg
	return &s, nil
}

func (s *Server) Start(wg *sync.WaitGroup) {
	s.router.LoadHTMLGlob("internal/server/templates/*.tmpl")
	s.router.POST("/update/:type/:name/:value", handlers.NewUpdateMetricHandler(s.storage))
	s.router.POST("/update/", handlers.NewRestUpdateMetricHandler(s.storage, s.service, s.lg))
	s.router.POST("/value/", handlers.NewShowRestMetricHandler(s.storage, s.lg))
	s.router.GET("/value/:type/:name", handlers.NewShowMetricHandler(s.storage))
	s.router.GET("/ping", handlers.NewPingHandler(s.storage, s.lg))
	s.router.GET("/", handlers.NewRootHandler(s.storage))

	s.lg.DebugCtx(s.ctx, "start", zap.String("config", s.config.String()))

	server := &http.Server{
		Addr:              s.config.Address,
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           s.router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.lg.ErrorCtx(s.ctx, "Failed", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-s.ctx.Done()
		s.lg.DebugCtx(s.ctx, "Graceful shutdown")
		if err := server.Shutdown(s.ctx); err != nil {
			s.lg.ErrorCtx(s.ctx, "Shutdown failed", zap.Error(err))
		}
	}()
}

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
		)

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			lg.ErrorCtx(ctx, "read body", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		c.Set("request_id", requestID)
		c.Next()

		status := c.Writer.Status()
		bodySize := c.Writer.Size()

		lg.InfoCtx(ctx, "request",
			zap.String("body", string(body)),
			zap.Int("status", status),
			zap.Int("body_size", bodySize),
			zap.Duration("duration", time.Since(start)),
		)
	}
}
