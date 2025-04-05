package osexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is an analyzer that finds os.Exit() calls
var Analyzer = &analysis.Analyzer{
	Name: "os_exit",
	Doc:  "find os.Exit() calls that may cause unexpected program termination",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	result := pass.ResultOf[inspect.Analyzer]
	if result == nil {
		return nil, nil // Skip analysis if inspector is not available
	}

	inspector, ok := result.(*inspector.Inspector)
	if !ok {
		return nil, nil // Skip analysis if type assertion fails
	}

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
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
		(*ast.Ident)(nil),
	}

	inspector.Preorder(nodeFilter, func(n ast.Node) {
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
			if x.Name == "Exit" {
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

	return nil, nil
}
