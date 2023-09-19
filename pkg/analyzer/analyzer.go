package analyzer

import (
	"flag"
	"fmt"
	"go/ast"
	"go/types"

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
		FactTypes: []analysis.Fact{new(isSafeFact)},
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

		if !doesFuncContainRecover(fdecl.Body) {
			return
		}

		fn := pass.TypesInfo.ObjectOf(fdecl.Name)
		if fn == nil {
			// Type information may be incomplete.
			fmt.Printf("Adding safeFunction failed :( %s\n", fdecl.Name)

			return
		}

		pass.ExportObjectFact(fn, new(isSafeFact))
	})

	return nil
}

// doesFuncContainRecover checks if function has a recover.
func doesFuncContainRecover(blckStmt *ast.BlockStmt) bool {
	for _, stmt := range blckStmt.List {
		switch stmt := stmt.(type) {
		case *ast.DeferStmt:
			hasRecover := false
			// TODO: maybe refactor into a function
			ast.Inspect(stmt.Call.Fun, func(fnLitNode ast.Node) bool {
				if _, ok := fnLitNode.(*ast.GoStmt); ok {
					// Each Goroutine has it's recover handler, so we don't want to inspect it.
					return false
				}

				// TODO: should we force the recover the be at the top off the function?
				ident, ok := fnLitNode.(*ast.Ident)
				if !ok {
					return true
				}

				// track recovery
				hasRecover = hasRecover || ident.Name == "recover"
				return false
			})

			if hasRecover {
				return true
			}
		// TODO: maybe check if it's just a bunch of function calls, and each function has a recover than the function is safe
		default:
			fmt.Printf("[doesFuncContainRecover] Not handling type: %T\n", stmt)
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
		// TODO: instead go through function declaration, that way you can keep a Set of relevant assignments to get the relevant function literal assigned to the variable/selector being called.
		// Track assignment to func, slice of function, nested map, structs containing functions where last value is functions. For fields in struct, do I want to assume the variable doesn't get reassigned? Or, do I need to follow the function call to make sure the field doesn't get reassigned? What if some complicated logic checks for assignment e.g. if rand() < 100, make function nil?

		// Even for normal variable, if it's a pointer to a function literal it can change, Actually
		// ignore this, assume function literal will not change after initialized. You can create option
		// to always flag calls to Goroutine with anything other than direct literal or function declaration as an error. Mention it's faster and has less false negatives. But, has more
		// false positives.

		// TODO: ideally we can track assignment from by following function
		(*ast.GoStmt)(nil), /* Find Goroutines */
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		goStmt := node.(*ast.GoStmt)

		if !isFuncSafe(pass, goStmt.Call.Fun) {
			pass.Reportf(node.Pos(), "Goroutine should have a defer recover")
		}
	})

	return nil
}

func isFuncSafe(pass *analysis.Pass, node ast.Node) bool {
	switch fn := node.(type) {
	case *ast.FuncLit:
		return doesFuncContainRecover(fn.Body)
	case *ast.Ident:

		tfn := pass.TypesInfo.ObjectOf(fn)
		if tfn == nil {
			fmt.Printf("Missing function info: %s\n", fn.Name)
			return false
		}

		return pass.ImportObjectFact(tfn, &isSafeFact{})
	case *ast.IndexExpr, *ast.IndexListExpr:
		x := getIDFromIndexParam(fn)
		id, _ := x.(*ast.Ident)
		if id == nil {
			return false
		}

		return isFuncSafe(pass, id)
	case *ast.SelectorExpr:
		id := fn.Sel

		if clit, ok := fn.X.(*ast.CompositeLit); ok {
			if len(clit.Elts) == 0 {
				return isFuncSafe(pass, id) // If we have an empty composite literal, then the selector has to be a method.
			}

			// We want to treat anonymous types the same as name type, so we get the underlying type
			clType, ok := getUnderlyingType(pass, clit.Type)
			if !ok {
				return isFuncSafe(pass, id)
			}
			switch st := clType.(type) {
			case *types.Struct:
				switch structDecl := clit.Elts[0].(type) {
				case *ast.KeyValueExpr:
					// KeyValueExpr are structs declared in the following way
					// myStruct{ myKey: myValue, myOtherKey, myOtherValue}
					for i, elt := range clit.Elts {
						elt, ok := elt.(*ast.KeyValueExpr)
						if !ok {
							panic(fmt.Sprintf("elt type at index %d, did not match value at index 0. initialType: %T, curType: %T", i, structDecl, elt))
						}

						k := elt.Key
						kID, ok := k.(*ast.Ident)
						if !ok {
							fmt.Printf("Key type of Key-value expression in a struct declaration was not a Identifier%T\n", k)
							continue
						}

						if kID.Name == id.Name {
							return isFuncSafe(pass, elt.Value)
						}
					}
				default:
					// The other type of struct declaration e.g.
					// myStruct{ myValue, myOtherValue}
					i, matched := getMatchedFieldIndex(st, id)
					if !matched {
						fmt.Printf("Unmatched selector type %T", structDecl)
						return isFuncSafe(pass, id)
					}

					return isFuncSafe(pass, clit.Elts[i])
				}

				return isFuncSafe(pass, id)
			default:
				fmt.Printf("Unhandled composite literal declaration %T\n", clit.Type)
			}
		}

		return isFuncSafe(pass, id)
	default:
		fmt.Printf("Unknown type: %T\n", fn)
		return false
	}
}

func getUnderlyingType(pass *analysis.Pass, node ast.Expr) (types.Type, bool) {
	if id, ok := node.(*ast.Ident); ok {
		obj := pass.TypesInfo.ObjectOf(id)
		if obj == nil {
			fmt.Printf("Could not find type of composite literal: %q", id)
			return nil, false
		}

		idTypeNamed, ok := obj.Type().(*types.Named)
		if !ok {
			panic(fmt.Sprintf("Expected named object to be named: %q, but got type %T", id, obj.Type()))
		}

		return idTypeNamed.Underlying(), true
	}

	return pass.TypesInfo.TypeOf(node), true
}

func getMatchedFieldIndex(st *types.Struct, id *ast.Ident) (int, bool) {
	for i := 0; i < st.NumFields(); i++ {
		fd := st.Field(i)
		if id.Name == fd.Name() {
			return i, true
		}
	}

	return -1, false
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

type isSafeFact struct{} // =>  *types.Func f is a function that won't panic

func (*isSafeFact) AFact() {}

func (*isSafeFact) String() string {
	return "isSafe"
}

func (s *isSafeFact) GobDecode(data []byte) error {
	if string(data) != "isSafe" {
		return fmt.Errorf("invalid GOB data: %q", data)
	}

	s = &isSafeFact{}
	return nil
}

func (*isSafeFact) GobEncode() ([]byte, error) {
	return []byte("isSafe"), nil
}
