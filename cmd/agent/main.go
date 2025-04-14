package main

import (
	"context"
	"log"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	BuildVersion string = "N/A"
	BuildDate    string = "N/A"
	BuildCommit  string = "N/A"
)

func main() {
	cfg, err := config.NewConfig(nil)
	if err != nil {
		log.Fatal(err)
	}

	lg, err := logging.MustZapLogger(zapcore.Level(cfg.LogLevel))
	if err != nil {
		log.Fatal(err)
	}

	info(lg)

	ctx := context.Background()
	agent := agent.NewAgent(
		lg,
		cfg,
		storage.NewMemoryStorage(lg),
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
