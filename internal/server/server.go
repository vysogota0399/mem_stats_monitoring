package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/handlers"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/logger"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

type Server struct {
	config  Config
	router  *gin.Engine
	storage storage.Storage
}

type NewServerOption func(*Server)

func NewServer(c Config, storage storage.Storage) (*Server, error) {
	s := Server{
		config: c,
		router: gin.New(),
	}

	s.router.Use(
		gin.Recovery(),
		withLogger(),
	)

	s.storage = storage
	return &s, nil
}

func (s *Server) Start() error {
	s.router.LoadHTMLGlob("internal/server/templates/*.tmpl")
	s.router.POST("/update/:type/:name/:value", handlers.NewUpdateMetricHandler(s.storage))
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

func withLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method

		uuid.NewV4()
		var requestID string

		if reqID := c.Request.Header.Get("X-Request-ID"); reqID != "" {
			requestID = reqID
		} else {
			requestID = uuid.NewV4().String()
		}

		sugar := logger.Log.Sugar()
		sugar.Infof("[%s] %s %s %s", requestID, method, path, raw)
		c.Next()

		status := c.Writer.Status()
		bodySize := c.Writer.Size()

		sugar.Infof("[%s] Response %d %d (%v)", requestID, status, bodySize, time.Since(start))
	}
}
