# safegoroutine

safegoroutine is a linter that ensures every Goroutine is has a defer to catch it. This is required because recover handler are not inherited by child Goroutines in Go.

In it's current state I personally wouldn't use it, since it gives too many false positives.

Instead I would recommend creating a central function to launch Goroutine e.g. 

```go
// Go safely launches a Goroutine, so that a panic doesn't bring down the host.
func Go(f func()) {
  go func() {
	defer func() {
	  if err := recover(); err != nil {
		// TODO: handle error
	  }
	}()

	f()
  }()
}

// CtxGo safely launches a Goroutine, so that a panic doesn't bring down the host.
// The context is used to improve our observability.
func CtxGo(ctx context.Context, f func()) {
	go func() {
	defer func() {
	  if err := recover(); err != nil {
		// TODO: handle error
	  }
	}()
	f()
  }()
}
```


TODOs
Allow assignment settings
1. None: don't allow any variables in Goroutines
2. Maintain invariant: 
	1. For the variable (including struct, slice, maps, chan, and combinations (slice of struct)) Make sure everything put into it is always valid. So, we don't care about branched e.g. switch, select, for, if, etc. Because we just look at the actual assignments. We track all function literal assignments
3. Assume unchanged: Assume initial declaration isn't changed
	- Will this ever be valuable? 


Some more TODO
- cast literal into type to run method
	e.g `go customString(s).safe()`
- use chatGPT to generate some random Go code that has some Go routines
- used chained function call e.g. newMyStruct(x).someOtherFunc().safe()
- Built-in type cast: `go string("some string")` - Can just let this be marked as unsafe
- Try to unsafeFuction where the recover happens in a nested Goroutine
-  In nestedSafeFunc call
	- insert empty line
	- also call function with no body cause it's safe
- In follow improve definition of "safe" function
	- right now
		- if function body is empty
		- it starts with a recover
	- some ideas, brainstorm ways for these to cause panics
		- safe type assertions
		- assignments to normal variables (no slices, and maps cause they need to be vetted)
		- something that can be gated behind config
			- Allows append, assignment to non-nil map
				- can potentially cause panic from memory overflow
			- Some arithmetic if we can verify 
	- list of known panics
		- panic in function
		- index
			- memory is nil: e.g. nil function or nil field (z.someField)
		- slice: out of bound slice
		- conversion: e.g. slice to array
		- adding memory: e.g. adding to slice or map
		- divide: divide by zero
		- shift, add, subtract: going out of bound
			- shift: by a number too bug, or too small if negative
