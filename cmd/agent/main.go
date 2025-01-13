package main

import (
	"context"
	"log"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap/zapcore"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	lg, err := logging.MustZapLogger(zapcore.Level(cfg.LogLevel))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	agent.NewAgent(
		lg,
		cfg,
		storage.NewMemoryStorage(lg),
	).Start(ctx)
}
