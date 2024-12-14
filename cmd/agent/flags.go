package main

import (
	"flag"
)

const (
	defaultReportIntercal = 10
	defaultPollInterval   = 2
)

var (
	flagServerAddr     string
	flagReportInterval int64
	flagPollInterval   int64
)

func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&flagReportInterval, "r", defaultReportIntercal, "Report interval")
	flag.Int64Var(&flagPollInterval, "p", defaultPollInterval, "Poll interval")

	flag.Parse()
}
