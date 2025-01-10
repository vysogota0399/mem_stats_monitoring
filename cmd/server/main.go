package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap/zapcore"
)

func main() {
	run()
}

func run() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
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

	strg, err := storage.NewStorage(ctx, cfg, &wg, lg)
	if err != nil {
		log.Fatal(err)
	}

	s, err := server.NewServer(
		ctx,
		cfg,
		strg,
		service.New(strg),
		lg,
	)
	if err != nil {
		log.Fatal(err)
	}

	s.Start(&wg)

	wg.Wait()
}
