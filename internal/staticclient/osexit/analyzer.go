package osexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const diagnosticMsg = "os.Exit call found - this may cause unexpected program termination"

// Analyzer is an analyzer that finds os.Exit() calls
var Analyzer = &analysis.Analyzer{
	Name: "os_exit",
	Doc:  "find os.Exit() calls that may cause unexpected program termination",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, imp := range pass.Files {
		hasOsImport := false
		for _, i := range imp.Imports {
			if i.Path.Value == `"os"` || i.Path.Value == `'os'` {
				hasOsImport = true
				break
			}
		}

		if !hasOsImport {
			continue
		}

		ast.Inspect(imp, func(n ast.Node) bool {
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
