package pkg

import (
	. "fmt"
)

// safeFuncLiteral tests if a function literal with a recover passes the lint checks.
func safeFuncLiteral() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				Printf("recover: %v\n", r)
			}
		}()

		Println("This should pass because it has a recover")
	}()
}

// unsafeFuncLiteral is a function that has a goroutine without a defer recover.
func unsafeFuncLiteral() {
	go func() { // want `Goroutine should have a defer recover`
		potentiallyUnsafeCode()
	}()
}

// safeFuncLiteralUntrackedRecoverValue is a function that has a goroutine with a recover, but
// the recover value is not tracked.
func safeFuncLiteralUntrackedRecoverValue() {
	go func() {
		defer func() {
			// Bad practice, but linter's job is just to make sure panics don't bring down the host
			recover()
		}()

		Println("This should pass because it has a recover")
	}()
}

// deferGoroutine is a function that has a defer statement without a recover.
func deferGoroutine() {
	go func() { // want `Goroutine should have a defer recover`
		defer func() {
			Println("Deferred but didn't recover :(")
		}()

		potentiallyUnsafeCode()
	}()
}

// nestedSafeFunc is a function that has a goroutine that starts another goroutine with a recover.
func nestedSafeFunc() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				Printf("recover: %v\n", r)
			}
		}()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					Printf("recover: %v\n", r)
				}
			}()

			Println("This should pass because it has a recover, even though it's nested")
		}()

		Println("This should pass because it has a recover")
	}()
}

// nestedUnsafeFunc is function that unsafely starts a safe Goroutine
func nestedUnsafeFunc() {
	go func() { // want `Goroutine should have a defer recover`
		potentiallyUnsafeCode()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					Printf("recover: %v\n", r)
				}
			}()

			Println("This should pass because it has a recover, even though it's nested")
		}()
	}()
}
