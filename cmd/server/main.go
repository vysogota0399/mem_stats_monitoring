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
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/logger"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
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

	if err := logger.Initialize(cfg.LogLevel, cfg.AppEnv); err != nil {
		log.Fatal(err)
	}

	storage, err := storage.NewPersistentMemory(ctx, cfg, &wg)
	if err != nil {
		log.Fatal(err)
	}

	s, err := server.NewServer(
		ctx,
		cfg,
		storage,
		service.New(storage),
	)
	if err != nil {
		log.Fatal(err)
	}

	s.Start(&wg)

	wg.Wait()
	logger.Log.Info("Main finished")
}
