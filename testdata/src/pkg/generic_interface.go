package pkg

type someGenericInterface[T, S any] interface {
	safe()
	unsafe()
}

// TODO: fix test
// safeMethodInGenericInterface is a function that has a method with a recover.
// func safeMethodInGenericInterface() {
// 	go someGenericInterface[any, any](myGenericStruct[any, any]{}).safe()
// 	go someGenericInterface[any, any](new(myGenericStruct[any, any])).safe()
// 	go someGenericInterface[any, any](newMyGenericStruct[any, any]()).safe()
// 	go someGenericInterface[any, any](newMyGenericStruct[any, any]().clone().clone()).safe()
// }

// unsafeMethodInGenericInterface is a function that starts a Goroutine with unsafe methods.
func unsafeMethodInGenericInterface() {
	go someGenericInterface[any, any](myGenericStruct[any, any]{}).unsafe() // want `Goroutine should have a defer recover`

	go someGenericInterface[any, any](new(myGenericStruct[any, any])).unsafe() // want `Goroutine should have a defer recover`
}

// TODO: fix test
// safeMethodInGenericInterfaceAssignment runs safe Goroutines from methods from structs that
// are assigned to a variable
// func safeMethodInGenericInterfaceAssignment() {
// 	v := someGenericInterface[any, any](myGenericStruct[any, any]{})
// 	go v.safe()

// 	p := someGenericInterface[any, any](&myGenericStruct[any, any]{})
// 	go p.safe()
// }

// unsafeMethodInGenericInterfaceAssignment runs unsafe Goroutines from methods from structs that
// are assigned to a variable
func unsafeMethodInGenericInterfaceAssignment() {
	v := someGenericInterface[any, any](myGenericStruct[any, any]{})
	go v.unsafe() // want `Goroutine should have a defer recover`

	p := someGenericInterface[any, any](&myGenericStruct[any, any]{})
	go p.unsafe() // want `Goroutine should have a defer recover`
}
