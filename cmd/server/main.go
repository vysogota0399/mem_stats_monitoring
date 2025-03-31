package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

func main() {
	run()
}

func run() {
	ctx, cancel := context.WithCancel(context.Background())
	errg, ctx := errgroup.WithContext(ctx)
	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
		<-exit
		cancel()
	}()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	lg, err := logging.MustZapLogger(zapcore.Level(cfg.LogLevel))
	if err != nil {
		log.Fatal(err)
	}

	strg, err := storage.NewStorage(ctx, cfg, errg, lg)
	if err != nil {
		log.Fatal(err)
	}

	s := server.NewServer(
		ctx,
		cfg,
		strg,
		service.New(strg),
		lg,
	)

	s.Start(errg)

	if err := errg.Wait(); err != nil {
		lg.FatalCtx(ctx, err.Error())
	}
}
