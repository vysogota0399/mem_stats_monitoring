package main

import (
	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/handlers"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storages/dump"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/crypto"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/fx"
)

var (
	BuildVersion string = "N/A"
	BuildDate    string = "N/A"
	BuildCommit  string = "N/A"
)

func main() {
	fx.New(CreateApp()).Run()
}

func CreateApp() fx.Option {
	return fx.Options(
		fx.Provide(
			config.NewConfig,
			logging.NewZapLogger,

			server.NewHTTPServer,

			fx.Annotate(crypto.NewDecryptor, fx.As(new(server.Decrypter))),
			fx.Annotate(config.NewFileConfig, fx.As(new(config.FileConfigurer))),
			fx.Annotate(server.NewRouter, fx.ParamTags(`group:"handlers"`)),

			fx.Annotate(storages.NewPGConnectionOpener, fx.As(new(storages.ConnectionOpener))),
			fx.Annotate(storages.NewGooseMigrator, fx.As(new(storages.Migrator))),
			fx.Annotate(storages.NewStorage, fx.As(new(storages.Storage))),
			fx.Annotate(dump.NewMetricsDumper, fx.As(new(storages.Dumper))),
			fx.Annotate(storages.NewStorage, fx.As(new(storages.IStorageTarget))),

			fx.Annotate(repositories.NewCounterRepository,
				fx.As(new(handlers.RootCounterRepository)),
				fx.As(new(handlers.IShowMetricCounterRepository)),
				fx.As(new(handlers.IShowRestMetricCounterRepository)),
				fx.As(new(service.CreateCntrRep)),
				fx.As(new(service.CntrRep)),
			),
			fx.Annotate(repositories.NewGaugeRepository,
				fx.As(new(handlers.RootGaugeRepository)),
				fx.As(new(handlers.IShowMetricGaugeRepository)),
				fx.As(new(handlers.IShowRestMetricGaugeRepository)),
				fx.As(new(service.CreateGaugeRep)),
				fx.As(new(service.GGRep)),
			),
			fx.Annotate(service.NewUpdateMetricService, fx.As(new(handlers.IUpdateMetricService))),
			fx.Annotate(service.NewUpdateMetricService, fx.As(new(handlers.IUpdateRestMetricService))),
			fx.Annotate(service.NewUpdateMetricsService, fx.As(new(handlers.IUpdateMetricsService))),

			AsHandlers(handlers.NewPingHandler),
			AsHandlers(handlers.NewRootHandler),
			AsHandlers(handlers.NewShowMetricHandler),
			AsHandlers(handlers.NewShowRestMetricHandler),
			AsHandlers(handlers.NewUpdateMetricHandler),
			AsHandlers(handlers.NewUpdateRestMetricHandler),
			AsHandlers(handlers.NewUpdatesRestMetricsHandler),
		),
		fx.Invoke(startHTTPServer),
	)
}

func startHTTPServer(*server.HTTPServer) {}

func AsHandlers(f any, ants ...fx.Annotation) any {
	ants = append(ants, fx.ResultTags(`group:"handlers"`))
	ants = append(ants, fx.As(new(server.Handler)))

	return fx.Annotate(
		f,
		ants...,
	)
}
