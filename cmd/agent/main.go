package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/clients"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/clients/grpc"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

var (
	BuildVersion string = "N/A"
	BuildDate    string = "N/A"
	BuildCommit  string = "N/A"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		os.Kill,
	)
	defer stop()

	cfg, err := config.NewConfig(config.NewFileConfig())
	if err != nil {
		log.Fatal(err)
	}

	lg, err := logging.MustZapLogger(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	info(lg)

	rep := agent.NewMetricsRepository(storage.NewMemoryStorage(lg))

	adapter, err := NewAdapter(ctx, &cfg, lg, rep)
	if err != nil {
		log.Fatal(err)
	}

	agent := agent.NewAgent(
		lg,
		cfg,
		rep,
		adapter,
	)

	agent.Start(ctx)
}

func info(lg *logging.ZapLogger) {
	lg.InfoCtx(context.Background(), "Build info",
		zap.String("version", BuildVersion),
		zap.String("date", BuildDate),
		zap.String("commit", BuildCommit),
	)
}

func NewAdapter(ctx context.Context, cfg *config.Config, lg *logging.ZapLogger, rep *agent.MetricsRepository) (agent.Adapter, error) {
	if cfg.GRPCPort != "" {
		rep, err := grpc.NewReporter(ctx, cfg, rep, lg)
		if err != nil {
			lg.ErrorCtx(ctx, "Failed to create grpc reporter", zap.Error(err))
			return nil, err
		}
		return rep, nil
	}

	return clients.NewCompReporter(cfg.ServerURL, lg, cfg, clients.NewDefaulut(), clients.NewIpSetter(lg), rep), nil
}
