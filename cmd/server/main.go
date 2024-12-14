package main

import (
	"flag"
	"log"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

func main() {
	parseFlags()
	run()
}

func run() {
	flag.Parse()

	config := server.NewConfig(flagRunAddr)
	s, err := server.NewServer(config, storage.New())
	if err != nil {
		log.Fatal(err)
	}

	if err := s.Start(); err != nil {
		panic(err)
	}
}
