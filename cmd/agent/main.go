package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/clients"
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

	agent := agent.NewAgent(
		lg,
		cfg,
		storage.NewMemoryStorage(lg),
		clients.NewCompReporter(cfg.ServerURL, lg, &cfg, clients.NewDefaulut()),
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
