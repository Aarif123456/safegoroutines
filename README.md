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

