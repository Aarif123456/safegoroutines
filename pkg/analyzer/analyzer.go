package analyzer

import (
	"flag"
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

//nolint:gochecknoglobals
var flagSet flag.FlagSet

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "safegoroutines",
		Doc:      "linter that ensures every Goroutine is has a defer to catch it. This is required because recover handler are not inherited by child Goroutines in Go",
		Run:      run,
		Flags:    flagSet,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

func run(pass *analysis.Pass) (any, error) {
	inspector, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, fmt.Errorf("Expected inspect.Analyzer to be an *inspector.Inspector, but got %T", pass.ResultOf[inspect.Analyzer])
	}

	nodeFilter := []ast.Node{
		(*ast.GoStmt)(nil), /* Find Goroutines */
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		hasRecover := false
		goStmt := node.(*ast.GoStmt)
		switch fn := goStmt.Call.Fun.(type) {
		case *ast.FuncLit:
			ast.Inspect(fn, func(fnLitNode ast.Node) bool {
				// TODO: should we force the recover the be at the top off the function?
				ident, ok := fnLitNode.(*ast.Ident)
				if !ok {
					return true
				}

				// track recovery
				hasRecover = hasRecover || ident.Name == "recover"
				return !hasRecover
			})
		case *ast.Ident:
			// fmt.Printf("Ident: %s\n", fn.Name)
			// TODO: create facts about the function that check if the functions start with a defer
		default:
			fmt.Printf("Unknown type: %T\n", goStmt.Call.Fun)
		}
		// TODO: don't handle just literals, we should be able to handle calls to Goroutine

		if !hasRecover {
			pass.Reportf(node.Pos(), "Goroutine should have a defer recover")
		}
	})

	return nil, nil
}
