// Package osexit provides a static analysis tool that detects calls to os.Exit() in Go code.
//
// The analyzer identifies direct calls to os.Exit() which can cause unexpected program termination
// and may prevent proper cleanup or resource management. This is particularly important in
// long-running services where graceful shutdown is essential.
//
// Usage:
//   - Import this package in your analysis tool
//   - Add the Analyzer to your analysis pass
//   - The analyzer will report diagnostics for any os.Exit() calls found
//
// Example diagnostic:
//
//	os.Exit call found - this may cause unexpected program termination
package osexit

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const diagnosticMsg = "os.Exit call found - this may cause unexpected program termination"

// Analyzer is an analyzer that finds os.Exit() calls
var Analyzer = &analysis.Analyzer{
	Name: "os_exit",
	Doc:  "find os.Exit() calls that may cause unexpected program termination",
	Run:  run,
}

// isIgnoredPath checks if the file path should be ignored by the analyzer
func isIgnoredPath(path string) bool {
	ignoredDirs := []string{
		".cache",
		"go-build",
		"vendor",
	}

	path = filepath.ToSlash(path)

	for _, dir := range ignoredDirs {
		if strings.Contains(path, dir) {
			return true
		}
	}

	return false
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Skip files in ignored directories
		if isIgnoredPath(pass.Fset.File(file.Pos()).Name()) {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			if n, ok := n.(*ast.File); ok {
				return n.Name.String() == "main"
			}

			switch x := n.(type) {
			case *ast.CallExpr:
				if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
						pass.Report(analysis.Diagnostic{
							Pos:     x.Pos(),
							Message: diagnosticMsg,
						})
					}
				}
			}

			return true
		})
	}

	return nil, nil
}
