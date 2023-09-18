package pkg

import (
	"errors"
	. "fmt"
)

// TODO: mark functions with no body as safe and add to test cases.

// potentiallyUnsafeCode represents some Go code, that can panic.
func potentiallyUnsafeCode() {
	Println("Some code that could potentially panic runs here...")
}

// funcWithRecover is a function that has a recovery handler.
func funcWithRecover() { // want funcWithRecover:`isSafe`
	defer func() {
		if r := recover(); r != nil {
			Printf("recover: %v\n", r)
		}
	}()

	Println("This should pass because it has a recover")
}

// safeFunc starts a Goroutines with a safe functions.
func safeFunc() {
	go funcWithRecover()

	// TODO: get this to pass
	// go funcWithOnlySafeCalls()
}

// unsafeFunc starts a Goroutines with an unsafe functions.
func unsafeFunc() {
	go potentiallyUnsafeCode() // want `Goroutine should have a defer recover`

	go funcWithMixedCalls() // want `Goroutine should have a defer recover`
}

// unsafeShadowedFunc is function that shadows a safe function with an unsafe function.
func unsafeShadowedFunc() {
	// We shadow a safe function with an unsafe function
	safeFunc := func() {
		potentiallyUnsafeCode()
	}

	go safeFunc() // want `Goroutine should have a defer recover`
}

// funcWithOnlySafeCalls is a function composed purely of safe function
func funcWithOnlySafeCalls() {
	funcWithRecover()
	funcWithRecover()
	funcWithRecover()
	funcWithRecover()
	funcWithRecover()
}

func funcWithMixedCalls() {
	funcWithRecover()
	potentiallyUnsafeCode()
	funcWithRecover()
}

func callToExternalPackge() {
	go errors.New("some error") // want `Goroutine should have a defer recover`
}

// TODO fix test
// func safeFuncShadow() {
// 	// We shadow the function because it can cause issues
// 	safeFunc := func() {
// 		defer func() {
// 			if r := recover(); r != nil {
// 				Printf("recover: %v\n", r)
// 			}
// 		}()

// 		Println("This should pass because it has a recover")
// 	}

// 	go safeFunc()
// }

// TODO: fix tst
// func safeFuncShadowsUnsafe() {
// 	// We shadow an unsafe function with a safe function
// 	potentiallyUnsafeCode := func() {
// 		defer func() {
// 			if r := recover(); r != nil {
// 				Printf("recover: %v\n", r)
// 			}
// 		}()

// 		Println("This should pass because it has a recover")
// 	}

// 	go potentiallyUnsafeCode()
// }
