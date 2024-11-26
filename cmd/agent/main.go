package main

import "github.com/vysogota0399/mem_stats_monitoring/internal/agent"

func main() {
	agent.NewAgent().Start()
}
