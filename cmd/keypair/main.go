package main

import (
	"context"
	"log"

	"github.com/vysogota0399/mem_stats_monitoring/internal/keypair"
	"github.com/vysogota0399/mem_stats_monitoring/internal/keypair/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("init config error: %s", err)
	}

	lg, err := logging.MustZapLogger(cfg)
	if err != nil {
		log.Fatalf("init logger error: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := keypair.NewGenerator(cfg, lg).Call(ctx); err != nil {
		lg.ErrorCtx(ctx, "error: %s", zap.Error(err))
	}

}
