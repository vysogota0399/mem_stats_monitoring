package server

import (
	"net/http"

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

func NewServer(c Config, options ...NewServerOption) (*Server, error) {
	s := Server{
		config: c,
		router: gin.Default(),
	}

	for _, opt := range options {
		opt(&s)
	}

	if s.logger == nil {
		s.logger = utils.InitLogger("[server]")
	}

	s.storage = storage.New()
	return &s, nil
}

func (s *Server) Start() error {
	s.router.LoadHTMLGlob("internal/server/templates/*.tmpl")
	s.router.POST("/update/:type/:name/:value", handlers.NewUpdateMetricHandler(s.storage, s.logger))
	s.router.GET("/value/:type/:name", handlers.NewShowMetricHandler(s.storage, s.logger))
	s.router.GET("/", handlers.NewRootHandler(s.storage, s.logger))
	if err := http.ListenAndServe(s.config.address, s.router); err != nil {
		return err
	}

	return nil
}

func SerLogger(logger utils.Logger) NewServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}
