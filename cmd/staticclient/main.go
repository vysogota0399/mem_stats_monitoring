package main

import (
	"context"
	"log"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/staticclient/multicheck"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

	mcheck, err := multicheck.NewMultiCheck(lg)
	if err != nil {
		lg.ErrorCtx(context.Background(), "Failed to create multicheck", zap.Error(err))
		return
	}
	if err := mcheck.Call(); err != nil {
		lg.ErrorCtx(context.Background(), "Failed to call multicheck", zap.Error(err))
		return
	}
}
