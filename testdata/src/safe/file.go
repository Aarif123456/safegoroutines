package safe

import (
	. "fmt"
)

func basicSafeFuncLiteral() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				Printf("recover: %v\n", r)
			}
		}()

		Println("This should pass because it has a recover")
	}()
}

func basicSafeFuncLiteralNoCatch() {
	go func() {
		defer func() {
			// Bad practice, but linter's job is just to make sure panics don't bring down the host
			recover()
		}()

		Println("This should pass because it has a recover")
	}()
}

func basicSafeFunc() {
	go safeFunc()
}

func basicNestedSafeFunc() {
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

func safeFunc() {
	defer func() {
		if r := recover(); r != nil {
			Printf("recover: %v\n", r)
		}
	}()

	Println("This should pass because it has a recover")
}
