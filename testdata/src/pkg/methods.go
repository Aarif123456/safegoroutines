package pkg

import (
	. "fmt"
)

type myStruct struct {
	f func()
}

func (m myStruct) safe() { // want safe:`isSafe`
	defer func() {
		if r := recover(); r != nil {
			Printf("recover: %v\n", r)
		}
	}()

	Println("This should pass because it has a recover")
}

func (m myStruct) unsafe() {
	Println("This should fail because it has no recover")
}

func (m myStruct) clone() myStruct {
	return m
}

func newMyStruct() myStruct {
	return myStruct{}
}

// safeMethod is a function that has a method with a recover.
func safeMethod() {
	go myStruct{}.safe()

	go new(myStruct).safe()
}

// unsafeMethod is a function that starts a Goroutine with unsafe methods.
func unsafeMethod() {
	go myStruct{}.unsafe() // want `Goroutine should have a defer recover`

	go new(myStruct).unsafe() // want `Goroutine should have a defer recover`
}

// TODO: fix test
// safeMethodAssignment starts safe Goroutines from methods from structs initialized to a struct.
// func safeMethodAssignment() {
// 	v := myStruct{}
// 	go v.safe()

// 	p := &myStruct{}
// 	go p.safe()
// }

// unsafeMethodAssignment starts unsafe Goroutines from methods from structs initialized to a variable.
func unsafeMethodAssignment() {
	v := myStruct{}
	go v.unsafe() // want `Goroutine should have a defer recover`

	p := &myStruct{}
	go p.unsafe() // want `Goroutine should have a defer recover`
}

// safeMethodExpression is a function that starts safe Goroutines with method expressions.
func safeMethodExpression() {
	go myStruct.safe(myStruct{})

	go (*myStruct).safe(&myStruct{})
}

// unsafeMethodExpression is a function that starts unsafe Goroutines with method expression.
func unsafeMethodExpression() {
	go myStruct.unsafe(myStruct{}) // want `Goroutine should have a defer recover`

	go (*myStruct).unsafe(&myStruct{}) // want `Goroutine should have a defer recover`
}

// safeFields is function that runs goroutines using the fields in a struct
func safeFields() {
	go myStruct{f: funcWithRecover}.f()

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
	go myStruct{f: potentiallyUnsafeCode}.f() // want `Goroutine should have a defer recover`
	go myStruct{}.f()                         // want `Goroutine should have a defer recover`
	go new(myStruct).f()                      // want `Goroutine should have a defer recover`

	go struct{ f func() }{f: potentiallyUnsafeCode}.f() // want `Goroutine should have a defer recover`
	go struct{ f func() }{}.f()                         // want `Goroutine should have a defer recover`
}

// TODO: fix test
// safeFieldsAssignment is function that runs goroutines using the fields from structs initialized to
// a variable.
// func safeFieldsAssignment() {
// 	v := myStruct{f: funcWithRecover}
// 	go v.f()

// 	p := &myStruct{f: funcWithRecover}
// 	go p.f()
// }

// unsafeFieldsAssignment is a function that starts a Goroutine with unsafe fields from structs
// initialized to a variable.
func unsafeFieldsAssignment() {
	v := myStruct{}
	go v.f() // want `Goroutine should have a defer recover`

	v1 := myStruct{f: potentiallyUnsafeCode}
	go v1.f() // want `Goroutine should have a defer recover`

	p := &myStruct{}
	go p.f() // want `Goroutine should have a defer recover`

	p1 := &myStruct{f: potentiallyUnsafeCode}
	go p1.f() // want `Goroutine should have a defer recover`
}
