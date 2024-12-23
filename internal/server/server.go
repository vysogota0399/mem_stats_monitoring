package server

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/handlers"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/logger"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

type Server struct {
	config  Config
	router  *gin.Engine
	storage storage.Storage
	service *service.Service
}

type NewServerOption func(*Server)

func NewServer(c Config, storage storage.Storage, service *service.Service) (*Server, error) {
	s := Server{
		config:  c,
		router:  gin.New(),
		service: service,
	}

	s.router.Use(
		gin.Recovery(),
		httpLogger(),
		gzip.Gzip(gzip.DefaultCompression),
	)

	s.storage = storage
	return &s, nil
}

func (s *Server) Start() error {
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

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func httpLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
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

		sugar.Infof("[%s] %s %s %s %s", requestID, method, path, raw, string(body))
		c.Next()

		status := c.Writer.Status()
		bodySize := c.Writer.Size()

		sugar.Infof("[%s] Response %d %d (%v)", requestID, status, bodySize, time.Since(start))
	}
}
