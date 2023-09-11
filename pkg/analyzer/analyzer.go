package analyzer

import (
	"flag"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

//nolint:gochecknoglobals
var flagSet flag.FlagSet

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:  "safegoroutines",
		Doc:   "linter that ensures every Goroutine is has a defer to catch it. This is required because recover handler are not inherited by child Goroutines in Go",
		Run:   run,
		Flags: flagSet,
	}
}

func run(pass *analysis.Pass) (any, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.GoStmt)(nil), /* Find Goroutines */
	}

	i.Preorder(nodeFilter, func(node ast.Node) {
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
				hasRecover = ident.Name == "recover"
				return !hasRecover
			})
		}
		// default:
		// 		// TODO: don't handle just literals, we should be able to handle calls to Goroutine
		if !hasRecover {
			pass.Reportf(node.Pos(), errorMsg)
		}
	})

	return nil, nil
}
