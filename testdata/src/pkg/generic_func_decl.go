package pkg

import (
	. "fmt"
)

// genericFunctionWithRecover is a generic function that has a recover.
func genericFunctionWithRecover[T any]() { // want genericFunctionWithRecover:`isSafe`
	defer func() {
		if r := recover(); r != nil {
			Printf("recover: %v\n", r)
		}
	}()

	Println("This should pass because it has a recover")
}

// genericFunctionMultipleParameterWithRecover is a generic function that has a recover.
func genericFunctionMultipleParameterWithRecover[T, S, V any]() { // want genericFunctionMultipleParameterWithRecover:`isSafe`
	defer func() {
		if r := recover(); r != nil {
			Printf("recover: %v\n", r)
		}
	}()

	Println("This should pass because it has a recover")
}

// potentiallyUnsafeGenericFunc represents some Go code, that can panic.
func potentiallyUnsafeGenericFunc[T any]() {
	Println("Some code that could potentially panic runs here...")
}

// safeGenericFunc is a function that starts a Goroutine with a safe function.
func safeGenericFunc() {
	go genericFunctionWithRecover[any]()
	go genericFunctionMultipleParameterWithRecover[any, any, any]()

	// TODO: get this to pass
	// go genericFuncWithOnlySafeCalls[any]()
}

// unsafeGenericFunc is a function that starts a Goroutine with a safe function.
func unsafeGenericFunc() {
	go potentiallyUnsafeGenericFunc[any]() // want `Goroutine should have a defer recover`

	go genericFuncWithMixedCalls[any]() // want `Goroutine should have a defer recover`
}

// TODO fix test
// func safeGenericFuncShadow() {
// 	// We shadow the function because it can cause issues
// 	safeGenericFunc := func() {
// 		defer func() {
// 			if r := recover(); r != nil {
// 				Printf("recover: %v\n", r)
// 			}
// 		}()

// 		Println("This should pass because it has a recover")
// 	}

// 	go safeGenericFunc()
// }

// unsafeShadowedGenericFunc is function that shadows a safe function with an unsafe function.
func unsafeShadowedGenericFunc() {
	// We shadow a safe function with an unsafe function
	safeGenericFunc := func() {
		potentiallyUnsafeGenericFunc[any]()
	}

	go safeGenericFunc() // want `Goroutine should have a defer recover`
}

// genericFuncWithOnlySafeCalls is a function composed purely of safe function
func genericFuncWithOnlySafeCalls[T any]() {
	genericFunctionWithRecover[T]()
	genericFunctionWithRecover[T]()
	genericFunctionWithRecover[T]()
	genericFunctionWithRecover[T]()
	genericFunctionWithRecover[T]()
}

// genericFuncWithMixedCalls is a function that calls some potentially unsafe functions.
func genericFuncWithMixedCalls[T any]() {
	genericFunctionWithRecover[T]()
	potentiallyUnsafeGenericFunc[T]()
	genericFunctionWithRecover[T]()
}
