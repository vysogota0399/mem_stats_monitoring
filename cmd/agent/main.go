package main

import (
	"time"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
)

func main() {
	parseFlags()
	config := agent.NewConfig(
		time.Duration(flagPollInterval)*time.Second,
		time.Duration(flagReportInterval)*time.Second,
		flagServerAddr,
	)

	agent.NewAgent(
		config,
		storage.NewMemoryStorage(),
	).Start()
}
