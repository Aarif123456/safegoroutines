package pkg

import (
	. "fmt"
)

func unsafeShadowedFunc() {
	// We shadow a safe function with an unsafe function
	safeFunc := func() {
		Println("This should fail because it can panic")
	}

	go safeFunc() // want `Goroutine should have a defer recover`
}

func unsafeFunc() {
	Println("This should fail because it can panic")
}
