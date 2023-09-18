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
		Name:      "safegoroutines",
		Doc:       "linter that ensures every Goroutine has a defer to catch it. This is required because recover handler are not inherited by child Goroutines in Go",
		Run:       run,
		Flags:     flagSet,
		Requires:  []*analysis.Analyzer{inspect.Analyzer},
		FactTypes: []analysis.Fact{new(isSafe)},
	}
}

func run(pass *analysis.Pass) (any, error) {
	if err := annotateSafeFunc(pass); err != nil {
		return nil, err
	}

	if err := validateGoroutines(pass); err != nil {
		return nil, err
	}

	return nil, nil
}

func annotateSafeFunc(pass *analysis.Pass) error {
	inspector, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return fmt.Errorf("Expected inspect.Analyzer to be an *inspector.Inspector, but got %T", pass.ResultOf[inspect.Analyzer])
	}

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil), /* Find Function */
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		fdecl, ok := node.(*ast.FuncDecl)
		if !ok || fdecl.Body == nil {
			return
		}

		if !doesFuncContainRecover(fdecl) {
			return
		}

		fn := pass.TypesInfo.ObjectOf(fdecl.Name)
		if fn == nil {
			// Type information may be incomplete.
			fmt.Printf("Adding safeFunction failed :( %s\n", fdecl.Name)

			return
		}

		pass.ExportObjectFact(fn, new(isSafe))
	})

	return nil
}

// doesFuncContainRecover checks if function has a recover.
func doesFuncContainRecover(fdecl *ast.FuncDecl) bool {
	for _, stmt := range fdecl.Body.List {
		hasRecover := false
		switch stmt := stmt.(type) {
		case *ast.DeferStmt:
			// TODO: maybe refactor into a function
			ast.Inspect(stmt.Call.Fun, func(fnLitNode ast.Node) bool {
				// TODO: should we force the recover the be at the top off the function?
				ident, ok := fnLitNode.(*ast.Ident)
				if !ok {
					return true
				}

				// track recovery
				hasRecover = hasRecover || ident.Name == "recover"
				return false
			})
		// TODO: maybe check if it's just a bunch of function calls, and each function has a recover than the function is safe
		default:
			fmt.Printf("[doesFuncContainRecover] Not handling type: %T\n", stmt)
		}

		if hasRecover {
			return true
		}
	}

	return false
}

func validateGoroutines(pass *analysis.Pass) error {
	inspector, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return fmt.Errorf("Expected inspect.Analyzer to be an *inspector.Inspector, but got %T", pass.ResultOf[inspect.Analyzer])
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

				if _, ok := fnLitNode.(*ast.GoStmt); ok {
					// Each Goroutine has it's recover handler, so we don't want to inspect it.
					return false
				}

				ident, ok := fnLitNode.(*ast.Ident)
				if !ok {
					return true
				}

				// track recovery
				hasRecover = hasRecover || ident.Name == "recover"
				return false
			})
		case *ast.Ident:
			tFn := pass.TypesInfo.ObjectOf(fn)
			if tFn == nil {
				fmt.Printf("Missing function info: %s\n", fn.Name)
				break
			}

			var fact isSafe
			hasRecover = pass.ImportObjectFact(tFn, &fact)
		case *ast.IndexExpr, *ast.IndexListExpr:
			x := getIDFromIndexParam(fn)
			id, _ := x.(*ast.Ident)
			if id == nil {
				break
			}

			tFn := pass.TypesInfo.ObjectOf(id)
			if tFn == nil {
				fmt.Printf("Missing function info: %s\n", id.Name)
				break
			}

			var fact isSafe
			hasRecover = pass.ImportObjectFact(tFn, &fact)
		default:
			fmt.Printf("Unknown type: %T\n", goStmt.Call.Fun)
		}
		// TODO: don't handle just literals, we should be able to handle calls to Goroutine

		if !hasRecover {
			pass.Reportf(node.Pos(), "Goroutine should have a defer recover")
		}
	})

	return nil
}

// TODO: the point of this function is to get the function that was called by Goroutine. If this isn't
// working I can manually handle these cases as well.

// goStmtFunc returns the ast.Node of a call expression
// that was invoked as a go statement. Currently, only
// function literals declared in the same function, and
// static calls within the same package are supported.
//
// Modified https://cs.opensource.google/go/x/tools/+/refs/tags/v0.13.0:go/analysis/passes/testinggoroutine/testinggoroutine.go;l=118;drc=255eeebbce77653b04b0d73a4d5f9436b38c8fdd
func goStmtFun(pass *analysis.Pass, goStmt *ast.GoStmt) ast.Node {
	switch fun := goStmt.Call.Fun.(type) {
	case *ast.IndexExpr, *ast.IndexListExpr:
		x := getIDFromIndexParam(fun)
		id, _ := x.(*ast.Ident)
		if id == nil {
			break
		}

		if id.Obj == nil {
			break
		}

		if funDecl, ok := id.Obj.Decl.(ast.Node); ok {
			return funDecl
		}
	case *ast.Ident:
		// TODO(cuonglm): improve this once golang/go#48141 resolved.
		if fun.Obj == nil {
			break
		}
		if funDecl, ok := fun.Obj.Decl.(ast.Node); ok {
			return funDecl
		}
	case *ast.FuncLit:
		return goStmt.Call.Fun
	}

	return goStmt.Call
}

func getIDFromIndexParam(n ast.Node) ast.Expr {
	switch e := n.(type) {
	case *ast.IndexExpr:
		return e.X
	case *ast.IndexListExpr:
		return e.X
	default:
		return nil
	}
}

type isSafe struct{} // =>  *types.Func f is a function that won't panic

func (*isSafe) AFact() {}

func (*isSafe) String() string {
	return "isSafe"
}

func (s *isSafe) GobDecode(data []byte) error {
	if string(data) != "isSafe" {
		return fmt.Errorf("invalid GOB data: %q", data)
	}

	s = &isSafe{}
	return nil
}

func (*isSafe) GobEncode() ([]byte, error) {
	return []byte("isSafe"), nil
}
