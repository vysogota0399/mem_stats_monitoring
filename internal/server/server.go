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
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/logger"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

type Server struct {
	config  config.Config
	router  *gin.Engine
	storage storage.Storage
	service *service.Service
	ctx     context.Context
}

type NewServerOption func(*Server)

func NewServer(ctx context.Context, c config.Config, storage storage.Storage, service *service.Service) (*Server, error) {
	s := Server{
		config:  c,
		router:  gin.New(),
		service: service,
		ctx:     ctx,
	}

	s.router.Use(
		gin.Recovery(),
		httpLogger(),
		gzip.Gzip(gzip.DefaultCompression),
	)

	s.storage = storage
	logger.Log.Sugar().Debugf("Config: %s", c)
	return &s, nil
}

func (s *Server) Start(wg *sync.WaitGroup) {
	s.router.LoadHTMLGlob("internal/server/templates/*.tmpl")
	s.router.POST("/update/:type/:name/:value", handlers.NewUpdateMetricHandler(s.storage))
	s.router.POST("/update/", handlers.NewRestUpdateMetricHandler(s.storage, s.service))
	s.router.POST("/value/", handlers.NewShowRestMetricHandler(s.storage))
	s.router.GET("/value/:type/:name", handlers.NewShowMetricHandler(s.storage))
	s.router.GET("/", handlers.NewRootHandler(s.storage))

	server := &http.Server{
		Addr:              s.config.Address,
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           s.router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Sugar().Infof("listen: %s\n", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				logger.Log.Sugar().Debugln("Gracefull shutdown Server")
				if err := server.Shutdown(s.ctx); err != nil {
					logger.Log.Sugar().Error(err)
				}

				return
			}
		}
	}()
}

func httpLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		sugar := logger.Log.Sugar()
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			sugar.Error(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		var requestID string
		if reqID := c.Request.Header.Get("X-Request-ID"); reqID != "" {
			requestID = reqID
		} else {
			requestID = uuid.NewV4().String()
		}

		sugar.Infow(
			"Request",
			"request_id", requestID,
			"method", method,
			"path", path,
			"params", c.Params,
			"body", string(body),
		)
		c.Next()

		status := c.Writer.Status()
		bodySize := c.Writer.Size()

		sugar.Infow(
			"Response",
			"request_id", requestID,
			"status", status,
			"body_size", bodySize,
			"duration", time.Since(start),
		)
	}
}
