package server

import (
	"context"
	"fmt"
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/fx"
)

type Handler interface {
	Registrate() (Route, error)
}

type Route struct {
	Handler       gin.HandlerFunc
	Path          string
	Method        string
	HTMLTemplates []*template.Template
}

type Router struct {
	router    *gin.Engine
	decrypter Decrypter
}

func NewRouter(handlers []Handler, lc fx.Lifecycle, lg *logging.ZapLogger, cfg *config.Config, decrypter Decrypter) *Router {
	r := &Router{router: gin.Default(), decrypter: decrypter}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				mws, err := middlewares(cfg, lg, decrypter)
				if err != nil {
					return fmt.Errorf("router: collect middlewares error %w", err)
				}

				r.router.Use(mws...)

				for _, handler := range handlers {
					route, err := handler.Registrate()
					if err != nil {
						return err
					}

					if route.Path == "" {
						lg.DebugCtx(ctx, "route path is empty")
						continue
					}

					if len(route.HTMLTemplates) > 0 {
						for _, tmp := range route.HTMLTemplates {
							r.router.SetHTMLTemplate(tmp)
						}
					}

					r.router.Handle(route.Method, route.Path, route.Handler)
				}

				return nil
			},
			OnStop: func(ctx context.Context) error {
				return nil
			},
		},
	)

	return r
}
