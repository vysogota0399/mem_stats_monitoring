package main

import (
	"flag"
	"log"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/logger"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

var flagRunAddr string

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()
}

func main() {
	parseFlags()
	run()
}

func run() {
	config := server.NewConfig(flagRunAddr)
	if err := logger.Initialize(config.LogLevel, config.AppEnv); err != nil {
		log.Fatal(err)
	}

	storage := storage.New()
	s, err := server.NewServer(
		config,
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
