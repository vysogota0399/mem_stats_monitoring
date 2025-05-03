// Package server handles the initialization and operation of the web server.
// It defines endpoints, handlers, and middleware for the metrics collection service.
package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// IServer defines the interface for server operations
type IServer interface {
	Serve(net.Listener) error
	Shutdown(context.Context) error
}

type HTTPServer struct {
	Router *Router
	srv    IServer
	cfg    *config.Config
	lg     *logging.ZapLogger
}

func NewHTTPServer(lc fx.Lifecycle, cfg *config.Config, r *Router, lg *logging.ZapLogger) *HTTPServer {
	s := &HTTPServer{
		srv:    &http.Server{Addr: cfg.Address, Handler: r.router, ReadHeaderTimeout: time.Minute},
		Router: r,
		cfg:    cfg,
		lg:     lg,
	}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				lg.InfoCtx(ctx, "Starting HTTP server", zap.Any("cfg", s.cfg))
				return s.Start(ctx)
			},
			OnStop: func(ctx context.Context) error {
				return s.Shutdown(ctx)
			},
		},
	)

	return s
}

func (s *HTTPServer) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return err
	}

	go func() {
		if err := s.srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.lg.ErrorCtx(ctx, "http_server: serve failer error", zap.Error(err))
		}
	}()

	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
