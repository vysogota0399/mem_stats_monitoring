package main

import (
	"github.com/vysogota0399/mem_stats_monitoring/internal/server"
)

func main() {
	run()
}

func run() {
	s, err := server.NewServer(server.NewConfig())
	if err != nil {
		panic(err)
	}

	if err := s.Start(); err != nil {
		panic(err)
	}
}
