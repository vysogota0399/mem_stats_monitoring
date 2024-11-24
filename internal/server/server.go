package server

import (
	"fmt"
	"net/http"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

type Server struct {
	config  Config
	mux     *http.ServeMux
	logger  utils.Logger
	storage storage.Storage
}

type NewServerOption func(*Server)

func NewServer(c Config, options ...NewServerOption) (*Server, error) {
	s := Server{
		config: c,
		mux:    http.NewServeMux(),
	}

	for _, opt := range options {
		opt(&s)
	}

	if s.logger == nil {
		s.logger = utils.InitLogger("[server]")
	}

	s.storage = storage.NewMemoryStorage()
	return &s, nil
}

func (s *Server) Start() error {
	s.logger.Printf("Start\n%v", s)
	s.mux.Handle(`/update/`, Conveyor(UpdateMetricHandler{logger: s.logger, app: s}, RequestLogger))

	if err := http.ListenAndServe(s.address(), s.mux); err != nil {
		return err
	}

	return nil
}

func SerLogger(logger utils.Logger) NewServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

func (s Server) String() string {
	return fmt.Sprintf("Host: %s\nPort: %d", s.config.Host, s.config.Port)
}
func (s *Server) address() string {
	return fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
}
