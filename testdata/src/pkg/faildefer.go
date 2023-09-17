package pkg

import (
	. "fmt"
)

func deferGoroutine() {
	go func() { // want `Goroutine should have a defer recover`
		defer func() {
			Println("Deferred but didn't recover :(")
		}()

		Println("This should pass because it has a recover")
	}()
}
