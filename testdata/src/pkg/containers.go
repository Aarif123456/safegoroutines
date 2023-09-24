package pkg

// TODO: try calling function stored in array and in map: e.g. a[0](); m[k]()
// TODO: If it's static map/slice and you can analyze it, maybe let it go. Otherwise, you should fail

// TODO: call function from slice in map m[k][0]()
// TODO: call function from map in slice a[0][k]()

// TODO: call function from struct in a slice a[0].f() and a[0].safe()
// TODO: call function from struct in a map m[k].f() and m[k].safe()

// TODO: call function from struct in a slice of map and map of slice

// TODO: try to shrink slice and then call function from it's element a[low:high][index]()

import (
	. "fmt"
)

// safeArrayCalls is a function that starts Goroutine from functions inside arrays.
func safeArrayCalls() {
	go []myStruct{
		{},
	}[0].safe()

	go []myStruct{
		{potentiallyUnsafeCode},
		{f: funcWithRecover},
	}[1].f()

	go []myStruct{
		{funcWithRecover},
		{f: potentiallyUnsafeCode},
	}[0].f()

	// Just randomly added a parentheses to throw off linter
	go ([]myStruct{
		{funcWithRecover},
		{f: potentiallyUnsafeCode},
	}[0]).f()

	go []func(){
		funcWithRecover,
	}[0]()

	go []struct {
		f func()
		g func()
	}{
		{
			f: funcWithRecover,
			g: potentiallyUnsafeCode,
		},
		{
			potentiallyUnsafeCode, /* f */
			funcWithRecover,       /* g */
		},
	}[0].f()

	go []struct {
		f func()
		g func()
	}{
		{
			f: funcWithRecover,
			g: potentiallyUnsafeCode,
		},
		{
			potentiallyUnsafeCode, /* f */
			funcWithRecover,       /* g */
		},
	}[1].g()

	go new(myArray).safe()
}

// unsafeArray is a function that starts a Goroutine with unsafe methods.
func unsafeArray() {
	go myArray{}.unsafe() // want `Goroutine should have a defer recover`

	go new(myArray).unsafe() // want `Goroutine should have a defer recover`
}

// TODO: fix test
// safeArrayAssignment starts safe Goroutines from methods from structs initialized to a struct.
func safeArrayAssignment() {
	v := myArray{}
	go v.safe()

	p := &myArray{}
	go p.safe()

	// TODO: use new to make array
	go new(myArray).safe()

	// TODO: add value to array via append
}

// unsafeArrayAssignment starts unsafe Goroutines from methods from structs initialized to a variable.
func unsafeArrayAssignment() {
	v := myArray{}
	go v.unsafe() // want `Goroutine should have a defer recover`

	p := &myArray{}
	go p.unsafe() // want `Goroutine should have a defer recover`
}

// safeArrayExpression is a function that starts safe Goroutines with method expressions.
func safeArrayExpression() {
	go myArray.safe(myArray{})

	go (*myArray).safe(&myArray{})
}

// unsafeArrayExpression is a function that starts unsafe Goroutines with method expression.
func unsafeArrayExpression() {
	go myArray.unsafe(myArray{}) // want `Goroutine should have a defer recover`

	go (*myArray).unsafe(&myArray{}) // want `Goroutine should have a defer recover`
}

// safeFields is function that runs goroutines using the fields in a struct
func safeFields() {
	go myArray{f: funcWithRecover}.f()

	go struct{ f func() }{f: funcWithRecover}.f()

	go struct {
		f func()
		g func()
	}{funcWithRecover, potentiallyUnsafeCode}.f()

	go struct {
		f func()
		g func()
	}{f: potentiallyUnsafeCode, g: funcWithRecover}.g()
}

// unsafeFields is a function that starts a Goroutine with unsafe fields.
func unsafeFields() {
	go myArray{f: potentiallyUnsafeCode}.f() // want `Goroutine should have a defer recover`
	go myArray{}.f()                         // want `Goroutine should have a defer recover`
	go new(myArray).f()                      // want `Goroutine should have a defer recover`

	go struct{ f func() }{f: potentiallyUnsafeCode}.f() // want `Goroutine should have a defer recover`
	go struct{ f func() }{}.f()                         // want `Goroutine should have a defer recover`
}

// TODO: fix test
// safeFieldsAssignment is function that runs goroutines using the fields from structs initialized to
// a variable.
// func safeFieldsAssignment() {
// 	v := myArray{f: funcWithRecover}
// 	go v.f()

// 	p := &myArray{f: funcWithRecover}
// 	go p.f()
// }

// unsafeFieldsAssignment is a function that starts a Goroutine with unsafe fields from structs
// initialized to a variable.
func unsafeFieldsAssignment() {
	v := myArray{}
	go v.f() // want `Goroutine should have a defer recover`

	v1 := myArray{f: potentiallyUnsafeCode}
	go v1.f() // want `Goroutine should have a defer recover`

	p := &myArray{}
	go p.f() // want `Goroutine should have a defer recover`

	p1 := &myArray{f: potentiallyUnsafeCode}
	go p1.f() // want `Goroutine should have a defer recover`
}

// TODO: add slices and maps ////////////////////////////////////////

// safeArray is a function that has a method with a recover.
func safeArray() {
	go myArray{}.safe()

	go new(myArray).safe()
}

// TODO: fix test
// safeArrayAssignment starts safe Goroutines from methods from structs initialized to a struct.
// func safeArrayAssignment() {
// 	v := myArray{}
// 	go v.safe()

// 	p := &myArray{}
// 	go p.safe()
// }

// unsafeArray is a function that starts a Goroutine with unsafe methods.
func unsafeArray() {
	go myArray{}.unsafe() // want `Goroutine should have a defer recover`

	go new(myArray).unsafe() // want `Goroutine should have a defer recover`
}

// unsafeArrayAssignment starts unsafe Goroutines from methods from structs initialized to a variable.
func unsafeArrayAssignment() {
	v := myArray{}
	go v.unsafe() // want `Goroutine should have a defer recover`

	p := &myArray{}
	go p.unsafe() // want `Goroutine should have a defer recover`
}

// safeArrayExpression is a function that starts safe Goroutines with method expressions.
func safeArrayExpression() {
	go myArray.safe(myArray{})

	go (*myArray).safe(&myArray{})
}

// unsafeArrayExpression is a function that starts unsafe Goroutines with method expression.
func unsafeArrayExpression() {
	go myArray.unsafe(myArray{}) // want `Goroutine should have a defer recover`

	go (*myArray).unsafe(&myArray{}) // want `Goroutine should have a defer recover`
}

// safeFields is function that runs goroutines using the fields in a struct
func safeFields() {
	go myArray{f: funcWithRecover}.f()

	go struct{ f func() }{f: funcWithRecover}.f()

	go struct {
		f func()
		g func()
	}{funcWithRecover, potentiallyUnsafeCode}.f()

	go struct {
		f func()
		g func()
	}{f: potentiallyUnsafeCode, g: funcWithRecover}.g()
}

// unsafeFields is a function that starts a Goroutine with unsafe fields.
func unsafeFields() {
	go myArray{f: potentiallyUnsafeCode}.f() // want `Goroutine should have a defer recover`
	go myArray{}.f()                         // want `Goroutine should have a defer recover`
	go new(myArray).f()                      // want `Goroutine should have a defer recover`

	go struct{ f func() }{f: potentiallyUnsafeCode}.f() // want `Goroutine should have a defer recover`
	go struct{ f func() }{}.f()                         // want `Goroutine should have a defer recover`
}

// TODO: fix test
// safeFieldsAssignment is function that runs goroutines using the fields from structs initialized to
// a variable.
// func safeFieldsAssignment() {
// 	v := myArray{f: funcWithRecover}
// 	go v.f()

// 	p := &myArray{f: funcWithRecover}
// 	go p.f()
// }

// unsafeFieldsAssignment is a function that starts a Goroutine with unsafe fields from structs
// initialized to a variable.
func unsafeFieldsAssignment() {
	v := myArray{}
	go v.f() // want `Goroutine should have a defer recover`

	v1 := myArray{f: potentiallyUnsafeCode}
	go v1.f() // want `Goroutine should have a defer recover`

	p := &myArray{}
	go p.f() // want `Goroutine should have a defer recover`

	p1 := &myArray{f: potentiallyUnsafeCode}
	go p1.f() // want `Goroutine should have a defer recover`
}
