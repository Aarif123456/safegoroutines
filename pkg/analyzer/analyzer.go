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
		tFn := pass.TypesInfo.ObjectOf(fn)
		if tFn == nil {
			fmt.Printf("Missing function info: %s\n", fn.Name)
			return false
		}

		if tFnFrom, ok := tFn.(*types.Var); ok {
			tFn = tFnFrom.Origin()
		}

		tFn, ok := getFunctionOrigin(tFn)
		if !ok {
			fmt.Printf("Ident did not map to function: %q, %T\n", fn.Name, tFn)
		}

		return pass.ImportObjectFact(tFn, &isSafeFact{})
	case *ast.IndexExpr, *ast.IndexListExpr:
		x := getIDFromIndexParam(fn)
		id, _ := x.(*ast.Ident)
		if id == nil {
			return false
		}

		return isFuncSafe(pass, id)
	case *ast.SelectorExpr:
		id := fn.Sel

		switch clit := fn.X.(type) {
		case *ast.CompositeLit:
			if len(clit.Elts) == 0 {
				return isFuncSafe(pass, id) // If we have an empty composite literal, then the selector has to be a method.
			}

			// We want to treat anonymous types the same as name type, so we get the underlying type
			clType, ok := getUnderlyingCompositeType(pass, clit)
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
			// TODO: handle slices, array and maps
			default:
				fmt.Printf("Unhandled composite literal declaration %T\n", clType)
			}
		case *ast.CallExpr:
			tClitFn, ok := pass.TypesInfo.TypeOf(clit.Fun).(*types.Named)
			if !ok {
				// TODO: try to get this line to print
				fmt.Printf("Call expression in composite literal has unexpected type: %q, want:*types.Named, got: %T\n", clit.Fun, pass.TypesInfo.TypeOf(clit.Fun))
				return isFuncSafe(pass, id)
			}

			switch tClitFn.Underlying().(type) {
			case *types.Interface:
				fmt.Printf("TODO: interface checking selector: %s, %+v, %T\n", id, clit.Args[0], clit.Args[0])
				// TODO: We need to get the underlying type of the interface by analyzing `clit.Args`
				// At the start, we can do a quick check to get the type only if a struct is passed in
				// myVar:= myStruct{}; (f(myVar))
				// i.e. . no chained function calls. Later, can try to handle interface being passed in as well
				return isFuncSafe(pass, id)
			default:
				fmt.Printf("Unhandled call expression before selector: %q, %T\n", tClitFn.Underlying(), tClitFn.Underlying())
			}
		case *ast.Ident:
			return isFuncSafe(pass, id)
		default:
			fmt.Printf("Unknown CompositeLit type: %T\n", clit)
		}

		return isFuncSafe(pass, id)
	default:
		fmt.Printf("Unknown type: %T\n", fn)
		return false
	}
}

// getUnderlyingCompositeType gets the underlying type associated with the [composite declaration].
// For example, for a struct e.g struct{}{}, myStructType{}, &myStruct{}, it'll return *types.Struct.
// The returned type must be either struct, array, slice, or map type
//
// [composite declaration]: https://go.dev/ref/spec#Composite_literals
func getUnderlyingCompositeType(pass *analysis.Pass, clit *ast.CompositeLit) (types.Type, bool) {
	node := clit.Type

	out := pass.TypesInfo.TypeOf(node)
	if out == nil {
		fmt.Printf("Could not find type of composite literal: %q", node)
		return nil, false
	}

	for {
		switch cur := out.(type) {
		default:
			fmt.Printf("getUnderlyingCompositeType returns invalid value %T", out)
			return out, false
		case *types.Struct, *types.Array, *types.Slice, *types.Map:
			return out, true
		case *types.Named:
			for cur.Origin() != nil && cur.Origin() != cur {
				// TODO: do I need this to get to the generic function original definition?
				// Also, what other cases am I handling by accident?
				cur = cur.Origin()
			}

			out = cur.Underlying()
			for out.Underlying() != nil && out.Underlying() != out {
				out = out.Underlying()
			}
		}
	}
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

// getFunctionOrigin gets the canonical Func i.e. the Func object recorded in Info.Defs.
func getFunctionOrigin(tFn types.Object) (types.Object, bool) {
	cur, ok := tFn.(*types.Func)
	if !ok {
		return tFn, false
	}

	for cur.Origin() != nil && cur.Origin() != cur {
		cur = cur.Origin()
	}

	return cur, true
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
