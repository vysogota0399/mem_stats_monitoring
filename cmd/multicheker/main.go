package main

import (
	"strings"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/vysogota0399/mem_stats_monitoring/internal/multicheker/osexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	checks := map[string]bool{
		"QF1010": true,
		"QF1003": true,
		"ST1023": true,
		"ST1018": true,
		"ST1016": true,
		"ST1013": true,
		"ST1012": true,
		"ST1011": true,
		"ST1006": true,
		"ST1005": true,
		"S1036":  true,
		"S1030":  true,
		"S1028":  true,
		"S1016":  true,
		"S1010":  true,
		"S1009":  true,
		"S1001":  true,
	}
	mychecks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		errcheck.Analyzer,
		osexit.Analyzer,
	}

	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") || checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}
