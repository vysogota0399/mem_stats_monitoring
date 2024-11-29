package main

import (
	"time"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent"
)

func main() {
	parseFlags()
	agent.NewAgent(
		agent.NewConfig(
			agent.SetPollInterval(time.Duration(flagPollInterval)*time.Second),
			agent.SetReportInterval(time.Duration(flagReportInterval)*time.Second),
			agent.SetServerURL(flagServerAddr),
		),
	).Start()
}
