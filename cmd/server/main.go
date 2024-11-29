package main

import (
	"flag"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
)

func main() {
	parseFlags()
	run()
}

func run() {
	flag.Parse()

	s, err := server.NewServer(
		server.NewConfig(
			server.SetAddress(flagRunAddr),
		),
	)
	if err != nil {
		panic(err)
	}

	if err := s.Start(); err != nil {
		panic(err)
	}
}
