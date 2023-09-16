package failbasic

import (
	. "fmt"
)

func basicUnsafeGoroutine() {
	go func() { // want `Goroutine should have a defer recover`
		Println("This should fail because it does not have a recover")
	}()
}
