// Package main provides a static analysis tool that combines multiple Go analyzers
// to perform comprehensive code analysis. The tool helps identify potential issues,
// bugs, and code quality problems in Go codebases.
//
// The tool includes the following analyzers:
//
// Built-in Go Analyzers:
//   - printf: Checks for incorrect usage of Printf-style formatting
//   - shadow: Detects shadowed variables
//   - structtag: Validates struct field tags
//   - errcheck: Ensures error return values are checked
//   - osexit: Detects usage of os.Exit() calls that may cause unexpected program termination
//
// Staticcheck Analyzers:
//
// Style Checks (ST):
//   - ST1023: Unnecessary type conversion
//   - ST1018: Unnecessary use of fmt.Sprintf
//   - ST1016: Unnecessary use of fmt.Sprint
//   - ST1013: Unnecessary use of fmt.Sprintln
//   - ST1012: Unnecessary use of fmt.Sprintf with no arguments
//   - ST1011: Unnecessary use of fmt.Sprintf with no arguments
//   - ST1006: Poorly chosen name for variable of type error
//   - ST1005: Poorly chosen name for variable of type error
//
// Simplification Checks (S):
//   - S1036: Unnecessary guard around call to delete
//   - S1030: Unnecessary use of fmt.Sprint
//   - S1028: Unnecessary use of fmt.Sprintf
//   - S1016: Use a constant instead of repeating the conversion
//   - S1010: Omit second value in type assertion
//   - S1009: Omit redundant nil check on ranged loop
//   - S1001: Use a simple channel send/receive instead of select with a single case
//
// Quick Fix Checks (QF):
//   - QF1010: Convert slice expression to use pointer to array
//   - QF1003: Expand call to fmt.Sprintf
//
// All SA (Static Analysis) Checks:
//
//	All analyzers starting with "SA" are enabled by default.
//
// Usage:
//
//	# Basic usage
//	staticclient ./...
//
//	# Analyze specific package
//	staticclient ./path/to/package
//
//	# Analyze multiple packages
//	staticclient ./pkg1/... ./pkg2/...
//
// Output:
//
//	The tool will output diagnostic messages for any issues found in the code.
//	Each diagnostic includes:
//	- File location
//	- Line number
//	- Issue description
//	- Suggested fixes (where applicable)
//
// Requirements:
//   - Go 1.16 or later
//   - Access to the required analyzer packages
//
// Dependencies:
//   - github.com/kisielk/errcheck
//   - github.com/vysogota0399/mem_stats_monitoring/internal/multicheker/osexit
//   - golang.org/x/tools/go/analysis
//   - golang.org/x/tools/go/analysis/passes/printf
//   - golang.org/x/tools/go/analysis/passes/shadow
//   - golang.org/x/tools/go/analysis/passes/structtag
//   - honnef.co/go/tools/staticcheck
package main

import (
	"strings"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/vysogota0399/mem_stats_monitoring/internal/staticclient/osexit"

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
