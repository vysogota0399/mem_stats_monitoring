// Package multicheck provides a comprehensive static analysis tool that combines multiple Go analyzers
// to perform code analysis. It helps identify potential issues, bugs, and code quality problems in Go codebases.
package multicheck

import (
	"context"
	"io"
	"os"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/vysogota0399/mem_stats_monitoring/internal/staticclient/osexit"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"gopkg.in/yaml.v2"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// Analyzer represents a configuration for a specific analyzer with its checks.
// It is used to configure which static analysis checks should be performed.
type Analyzer struct {
	// Name is the identifier of the analyzer (e.g., "staticcheck")
	Name string `yaml:"name"`
	// Checks is a list of specific checks to be performed by this analyzer
	Checks []string `yaml:"checks"`
}

// MultiCheck is the main struct that manages multiple analyzers and their configurations.
// It reads analyzer configurations from a YAML file and combines them with built-in analyzers.
type MultiCheck struct {
	// ExtraAnalyzers contains additional analyzers configured via YAML
	ExtraAnalyzers []Analyzer `yaml:"analyzers"`
	// lg is the logger instance used for logging operations
	lg *logging.ZapLogger
	// source is the reader for the configuration file
	source io.ReadCloser
	// skip is a flag to skip the multicheck
	skip bool
}

// NewMultiCheck creates a new MultiCheck instance with the specified logger.
// It attempts to read the configuration from "multicheck.yml" if it exists.
// Returns a MultiCheck instance and any error that occurred during initialization.
func NewMultiCheck(lg *logging.ZapLogger) (*MultiCheck, error) {
	mcheck := MultiCheck{
		ExtraAnalyzers: []Analyzer{},
		lg:             lg,
	}

	if _, err := os.Stat("multicheck_config.yml"); os.IsNotExist(err) {
		return &mcheck, nil
	}

	source, err := os.Open("multicheck.yml")
	if err != nil {
		lg.ErrorCtx(context.Background(), "Failed to open multicheck.yml", zap.Error(err))
		return &mcheck, nil
	}

	mcheck.source = source

	return &mcheck, nil
}

// Call executes all configured analyzers, including both built-in and extra analyzers.
// It combines the following analyzers:
// - Built-in Go analyzers (printf, shadow, structtag, errcheck, osexit)
// - Staticcheck analyzers
// - Style check analyzers
// - Simple analyzers
// - Quick fix analyzers
// Returns an error if any analyzer fails to execute.
func (m *MultiCheck) Call() error {
	if m.source != nil {
		err := yaml.NewDecoder(m.source).Decode(&m)
		if err != nil {
			m.lg.ErrorCtx(context.Background(), "Failed to decode multicheck.yml", zap.Error(err))
			return err
		}
	}

	usedAnalyzers := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		errcheck.Analyzer,
		osexit.Analyzer,
	}

	for _, scAnalyzer := range staticcheck.Analyzers {
		usedAnalyzers = append(usedAnalyzers, scAnalyzer.Analyzer)
	}

	allAnalyzers := make(map[string]*analysis.Analyzer)

	for _, lintAnalyzer := range stylecheck.Analyzers {
		allAnalyzers[lintAnalyzer.Analyzer.Name] = lintAnalyzer.Analyzer
	}

	for _, lintAnalyzer := range simple.Analyzers {
		allAnalyzers[lintAnalyzer.Analyzer.Name] = lintAnalyzer.Analyzer
	}

	for _, lintAnalyzer := range quickfix.Analyzers {
		allAnalyzers[lintAnalyzer.Analyzer.Name] = lintAnalyzer.Analyzer
	}

	for _, analyzer := range m.ExtraAnalyzers {
		for _, check := range analyzer.Checks {
			if _, ok := allAnalyzers[check]; ok {
				usedAnalyzers = append(usedAnalyzers, allAnalyzers[check])
			}
		}
	}

	m.lg.DebugCtx(context.Background(), "Running multicheck", zap.Int("analyzers", len(usedAnalyzers)))

	if !m.skip {
		multichecker.Main(usedAnalyzers...)
	}

	return nil
}
