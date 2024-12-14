package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/handlers"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

type Server struct {
	config  Config
	router  *gin.Engine
	logger  utils.Logger
	storage storage.Storage
}

type NewServerOption func(*Server)

func NewServer(c Config, storage storage.Storage) (*Server, error) {
	s := Server{
		config: c,
		router: gin.Default(),
	}

	if s.logger == nil {
		s.logger = utils.InitLogger("[server]")
	}

	s.storage = storage
	return &s, nil
}

func (s *Server) Start() error {
	s.logger.Println(s.config)
	s.router.LoadHTMLGlob("internal/server/templates/*.tmpl")
	s.router.POST("/update/:type/:name/:value", handlers.NewUpdateMetricHandler(s.storage, s.logger))
	s.router.GET("/value/:type/:name", handlers.NewShowMetricHandler(s.storage, s.logger))
	s.router.GET("/", handlers.NewRootHandler(s.storage, s.logger))

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
