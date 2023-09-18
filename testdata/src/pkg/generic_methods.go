package pkg

import (
	. "fmt"
)

type myGenericStruct[T, S any] struct {
	f func()
}

func (m myGenericStruct[T, S]) safe() { // want safe:`isSafe`
	defer func() {
		if r := recover(); r != nil {
			Printf("recover: %v\n", r)
		}
	}()

	Println("This should pass because it has a recover")
}

func (m myGenericStruct[T, S]) unsafe() {
	Println("This should fail because it has no recover")
}

// safeGenericMethod is a function that starts a Goroutine with safe generic methods.
func safeGenericMethod() {
	go myGenericStruct[any, any]{}.safe()

	go new(myGenericStruct[any, any]).safe()

}

// safeGenericMethodAssignment starts safe Goroutines from a generic struct assigned to a variable.
// func safeGenericMethodAssignment() {
// 	v := myGenericStruct[any, any]{}
// 	go v.safe()

// 	p := &myGenericStruct[any, any]{}
// 	go p.safe()
// }

// unsafeGenericMethod is a function that starts a Goroutine with unsafe generic methods.
func unsafeGenericMethod() {
	go myGenericStruct[any, any]{}.unsafe() // want `Goroutine should have a defer recover`

	go new(myGenericStruct[any, any]).unsafe() // want `Goroutine should have a defer recover`
}

// unsafeGenericMethodAssignment starts unsafe Goroutines from a generic struct assigned to a variable.
func unsafeGenericMethodAssignment() {
	v := myGenericStruct[any, any]{}
	go v.unsafe() // want `Goroutine should have a defer recover`

	p := &myGenericStruct[any, any]{}
	go p.unsafe() // want `Goroutine should have a defer recover`
}

// safeGenericFields is a function that runs goroutines using the fields in a generic struct.
func safeGenericFields() {
	go myGenericStruct[any, any]{f: funcWithRecover}.f()
	go struct{ f func() }{f: funcWithRecover}.f()
}

// TODO: fix test
// safeGenericFieldsAssignment tarts safe Goroutines using the fields in a generic struct, where
// the struct was assigned to a variable.
// func safeGenericFieldsAssignment() {
// 	v := myGenericStruct[any, any]{f: funcWithRecover}
// 	go v.f()

// 	p := &myGenericStruct[any, any]{f: funcWithRecover}
// 	go p.f()
// }

// unsafeGenericFields is a function that starts unsafe Goroutine using the fields in a generic struct.
func unsafeGenericFields() {
	go myGenericStruct[any, any]{f: potentiallyUnsafeCode}.f() // want `Goroutine should have a defer recover`
	go myGenericStruct[any, any]{}.f()                         // want `Goroutine should have a defer recover`
	go new(myGenericStruct[any, any]).f()                      // want `Goroutine should have a defer recover`

	go struct{ f func() }{f: potentiallyUnsafeCode}.f() // want `Goroutine should have a defer recover`
	go struct{ f func() }{}.f()                         // want `Goroutine should have a defer recover`
}

// unsafeGenericFieldsAssignment starts unsafe Goroutine using the fields in a generic struct,
// where the struct was assigned to a variable.
func unsafeGenericFieldsAssignment() {
	v := myGenericStruct[any, any]{}
	go v.f() // want `Goroutine should have a defer recover`

	v1 := myGenericStruct[any, any]{f: potentiallyUnsafeCode}
	go v1.f() // want `Goroutine should have a defer recover`

	p := &myGenericStruct[any, any]{}
	go p.f() // want `Goroutine should have a defer recover`

	p1 := &myGenericStruct[any, any]{f: potentiallyUnsafeCode}
	go p1.f() // want `Goroutine should have a defer recover`

}
