package osexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is an analyzer that finds os.Exit() calls
var Analyzer = &analysis.Analyzer{
	Name:     "osexit",
	Doc:      "find os.Exit() calls that may cause unexpected program termination",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Get the inspector result and check for nil
	result, ok := pass.ResultOf[inspect.Analyzer]
	if !ok || result == nil {
		// If inspector is not available, create a new one
		insp := inspector.New(pass.Files)
		runAnalysis(pass, insp)
		return nil, nil
	}

	insp, ok := result.(*inspector.Inspector)
	if !ok {
		// If type assertion fails, create a new inspector
		newInsp := inspector.New(pass.Files)
		runAnalysis(pass, newInsp)
		return nil, nil
	}

	runAnalysis(pass, insp)
	return nil, nil
}

func runAnalysis(pass *analysis.Pass, insp *inspector.Inspector) {
	// First, check if the os package is imported
	hasOsImport := false
	for _, imp := range pass.Files {
		for _, i := range imp.Imports {
			if i.Path.Value == `"os"` || i.Path.Value == `'os'` {
				hasOsImport = true
				break
			}
		}
	}

	if !hasOsImport {
		return
	}

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
		(*ast.Ident)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		switch x := n.(type) {
		case *ast.CallExpr:
			// Check direct os.Exit calls
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(x.Pos(), "os.Exit() call found - this may cause unexpected program termination")
				}
			}

		case *ast.Ident:
			// Check for os.Exit being assigned to a variable
			if x.Name == "Exit" && x.Obj != nil {
				if sel, ok := x.Obj.Decl.(*ast.ValueSpec); ok {
					if sel.Type != nil {
						if selExpr, ok := sel.Type.(*ast.SelectorExpr); ok {
							if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "os" {
								pass.Reportf(x.Pos(), "os.Exit being assigned to a variable - this may cause unexpected program termination")
							}
						}
					}
				}
			}
		}
	})
}
