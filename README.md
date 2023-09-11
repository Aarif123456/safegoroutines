# safegoroutine

safegoroutine is a linter that ensures every Goroutine is has a defer to catch it. This is required because recover handler are not inherited by child Goroutines in Go.
