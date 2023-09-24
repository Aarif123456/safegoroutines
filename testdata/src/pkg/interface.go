package pkg

type someInterface interface {
	safe()
	unsafe()
}

// TODO: fix test

// safeMethodInInterface is a function that has a method with a recover.
// func safeMethodInInterface() {
// 	go someInterface(myStruct{}).safe()
// 	go someInterface(newMyStruct()).safe()
// 	go someInterface(newMyStruct().clone().clone()).safe()

// 	go someInterface(new(myStruct)).safe()
// }

// unsafeMethodInInterface is a function that starts a Goroutine with unsafe methods.
func unsafeMethodInInterface() {
	go someInterface(myStruct{}).unsafe() // want `Goroutine should have a defer recover`

	go someInterface(new(myStruct)).unsafe() // want `Goroutine should have a defer recover`
}

// Fix test
// safeMethodInInterfaceAssignment starts safe Goroutines from structs casted to a interface.
// func safeMethodInInterfaceAssignment() {
// 	v := someInterface(myStruct{})
// 	go v.safe()

// 	p := someInterface(&myStruct{})
// 	go p.safe()
// }

// unsafeMethodInInterfaceAssignment is a function that starts a Goroutine with unsafe methods.
func unsafeMethodInInterfaceAssignment() {
	v := someInterface(myStruct{})
	go v.unsafe() // want `Goroutine should have a defer recover`

	p := someInterface(&myStruct{})
	go p.unsafe() // want `Goroutine should have a defer recover`
}
