package main

import (
	"log"

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
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err := logger.Initialize(cfg.LogLevel, cfg.AppEnv); err != nil {
		log.Fatal(err)
	}

	storage, err := storage.NewPersistentMemory(cfg)
	if err != nil {
		log.Fatal(err)
	}

	s, err := server.NewServer(
		cfg,
		storage,
		service.New(storage),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := s.Start(); err != nil {
		panic(err)
	}
}
