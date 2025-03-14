package utils

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"runtime"
	"sync"
)

var (
	// ReallyCrash controls the behavior of HandleCrash and defaults to
	// true. It's exposed so components can optionally set to false
	// to restore prior behavior. This flag is mostly used for tests to validate
	// crash conditions.
	ReallyCrash = true
)

// PanicHandlers is a list of functions which will be invoked when a panic happens.
var PanicHandlers = []func(context.Context, interface{}){logPanic}

type DoWorkPieceFunc func(piece int)

type options struct {
	chunkSize int
}

type Options func(*options)

// WithChunkSize allows to set chunks of work items to the workers, rather than
// processing one by one.
// It is recommended to use this option if the number of pieces significantly
// higher than the number of workers and the work done for each item is small.
func WithChunkSize(c int) func(*options) {
	return func(o *options) {
		o.chunkSize = c
	}
}

// ParallelizeUntil is a framework that allows for parallelizing N
// independent pieces of work until done or the context is canceled.
func ParallelizeUntil(ctx context.Context, workers, pieces int, doWorkPiece DoWorkPieceFunc, opts ...Options) {
	if pieces == 0 {
		return
	}
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}
	chunkSize := o.chunkSize
	if chunkSize < 1 {
		chunkSize = 1
	}

	chunks := ceilDiv(pieces, chunkSize)
	toProcess := make(chan int, chunks)
	for i := 0; i < chunks; i++ {
		toProcess <- i
	}
	close(toProcess)

	var stop <-chan struct{}
	if ctx != nil {
		stop = ctx.Done()
	}
	if chunks < workers {
		workers = chunks
	}
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer HandleCrash()
			defer wg.Done()
			for chunk := range toProcess {
				start := chunk * chunkSize
				end := start + chunkSize
				if end > pieces {
					end = pieces
				}
				for p := start; p < end; p++ {
					select {
					case <-stop:
						return
					default:
						doWorkPiece(p)
					}
				}
			}
		}()
	}
	wg.Wait()
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}

func HandleCrash(additionalHandlers ...func(interface{})) {
	if r := recover(); r != nil {
		additionalHandlersWithContext := make([]func(context.Context, interface{}), len(additionalHandlers))
		for i, handler := range additionalHandlers {
			handler := handler // capture loop variable
			additionalHandlersWithContext[i] = func(_ context.Context, r interface{}) {
				handler(r)
			}
		}
		handleCrash(context.Background(), r, additionalHandlersWithContext...)
	}
}

func handleCrash(ctx context.Context, r any, additionalHandlers ...func(context.Context, interface{})) {
	for _, fn := range PanicHandlers {
		fn(ctx, r)
	}
	for _, fn := range additionalHandlers {
		fn(ctx, r)
	}
	if ReallyCrash {
		// Actually proceed to panic.
		panic(r)
	}
}

// logPanic logs the caller tree when a panic occurs (except in the special case of http.ErrAbortHandler).
func logPanic(ctx context.Context, r interface{}) {
	if r == http.ErrAbortHandler {
		// honor the http.ErrAbortHandler sentinel panic value:
		//   ErrAbortHandler is a sentinel panic value to abort a handler.
		//   While any panic from ServeHTTP aborts the response to the client,
		//   panicking with ErrAbortHandler also suppresses logging of a stack trace to the server's error log.
		return
	}

	// Same as stdlib http server code. Manually allocate stack trace buffer size
	// to prevent excessively large logs
	const size = 64 << 10
	stacktrace := make([]byte, size)
	stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]

	// We don't really know how many call frames to skip because the Go
	// panic handler is between us and the code where the panic occurred.
	// If it's one function (as in Go 1.21), then skipping four levels
	// gets us to the function which called the `defer HandleCrashWithontext(...)`.
	//logger := logs

	// For backwards compatibility, conversion to string
	// is handled here instead of defering to the logging
	// backend.
	if _, ok := r.(string); ok {
		logs.Error(nil, "Observed a panic", "panic", r, "stacktrace", string(stacktrace))
	} else {
		logs.Error(nil, "Observed a panic", "panic", fmt.Sprintf("%v", r), "panicGoValue", fmt.Sprintf("%#v", r), "stacktrace", string(stacktrace))
	}
}
