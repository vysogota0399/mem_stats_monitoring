package main

import (
	"flag"
)

var (
	flagServerAddr     string
	flagReportInterval int64
	flagPollInterval   int64
)

func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&flagReportInterval, "r", 10, "Report interval")
	flag.Int64Var(&flagPollInterval, "p", 2, "Poll interval")

	flag.Parse()
}
