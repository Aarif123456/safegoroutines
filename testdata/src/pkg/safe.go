package pkg

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

func safeFuncShadow() {
	// We shadow the function because it can cause issues
	safeFunc := func() {
		defer func() {
			if r := recover(); r != nil {
				Printf("recover: %v\n", r)
			}
		}()

		Println("This should pass because it has a recover")
	}

	go safeFunc()
}

func safeFuncShadowsUnsafe() {
	// We shadow an unsafe function with a safe function
	unsafeFunc := func() {
		defer func() {
			if r := recover(); r != nil {
				Printf("recover: %v\n", r)
			}
		}()

		Println("This should pass because it has a recover")
	}

	go unsafeFunc()
}

func basicSafeFunc() {
	go safeFunc()
}

func safeFunc() { // want safeFunc:`isSafe`
	defer func() {
		if r := recover(); r != nil {
			Printf("recover: %v\n", r)
		}
	}()

	Println("This should pass because it has a recover")
}

// TODO: test safe method e.g. x := mystruct{}; x.f()
// TODO: test generic struct with safe method generictStruct[safeStruct]{}.f()
// TODO: test safe generic function e.g. f[safeStruct]()

// TODO: call something from another package e.g. errors.New

// TODO: declare anonymous type with function and call it
// type struct { f func()} { func() { TODO test if it's valid vs not valid } }.f()

// TODO: try method expression e.g. safeStruct.Test(safestruct{}, maybeSomeArgs)
// TODO: try method expression with pointer receiver e.g. (*safeStruct).Test(&safestruct{}, maybeSomeArgs)

// TODO: try calling function stored in array and in map: e.g. a[0](); m[k]()
// TODO: call function from slice in map m[k][0]()
// TODO: call function from map in slice a[0][k]()

// TODO: try to shrink slice and then call function from it's element a[low:high][index]()

// TODO: test function composed purely of safe function should be marked as safe
func safeNestedFunc() {
	safeFunc()
	safeFunc()
	safeFunc()
	safeFunc()
	safeFunc()
}
