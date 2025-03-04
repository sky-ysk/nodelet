package finisher

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/errors"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
	"net/http"
	goruntime "runtime"
	"time"
)

// ResultFunc 是一个返回 REST 结果的函数，可以在 goroutine 中运行
type ResultFunc func() (runtime.Object, error)

// result 存储 ResultFunc 函数的返回值或 panic
type result struct {
	// object 存储 ResultFunc 函数返回的响应
	object runtime.Object
	// err 存储 ResultFunc 函数返回的错误
	err error
	// reason 存储 ResultFunc 函数引发的 panic 的原因
	reason interface{}
}

// Return 处理 ResultFunc 函数返回的结果
func (r *result) Return() (runtime.Object, error) {
	switch {
	case r.reason != nil:
		// 执行 ResultFunc 的协程已经 panic，因此，将 panic 传播给调用者。
		panic(r.reason)
	case r.err != nil:
		return nil, r.err
	default:
		if status, ok := r.object.(*meta.Status); ok {
			//一个 api状态不为 success 的 Status 对象被视为“错误”，这会中断正常的响应流。
			if status.Status != meta.StatusSuccess {
				return nil, errors.FromObject(status)
			}
		}
		return r.object, nil
	}
}

// PostTimeoutLoggerFunc 是一个函数，可用于在请求超时后记录 ResultFunc 返回的结果。
// timedOutAt 是请求超时的时间。r 是子 goroutine 返回的结果。
type PostTimeoutLoggerFunc func(timedOutAt time.Time, r *result)

const (
	postTimeoutLoggerWait = 5 * time.Minute
)

// FinishRequest 使给定的 ResultFunc 异步并处理响应返回的错误。
func FinishRequest(ctx context.Context, fn ResultFunc) (runtime.Object, error) {
	return finishRequest(ctx, fn, postTimeoutLoggerWait, logPostTimeoutResult)
}

func finishRequest(ctx context.Context, fn ResultFunc, postTimeoutWait time.Duration, postTimeoutLogger PostTimeoutLoggerFunc) (runtime.Object, error) {
	resultCh := make(chan *result, 1)

	go func() {
		result := &result{}
		defer func() {
			reason := recover()
			if reason != nil {
				if reason != http.ErrAbortHandler {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:goruntime.Stack(buf, false)]
					reason = fmt.Sprintf("%v\n%s", reason, buf)
				}
				// 将 panic 原因存储到 result 中。
				result.reason = reason
			}
			// 将结果传播到父 goroutine
			resultCh <- result
		}()
		if object, err := fn(); err != nil {
			result.err = err
		} else {
			result.object = object
		}
	}()

	select {
	case result := <-resultCh:
		return result.Return()
	case <-ctx.Done():
		// 我们将向调用者发送一个 timeout 响应，但异步协程（sender）仍在执行 ResultFunc 函数。
		// 在这里启动一个协程（接收者）来等待 sender 发送结果，然后记录结果的详细信息。
		defer func() {
			go func() {
				timedOutAt := time.Now()
				var result *result
				select {
				case result = <-resultCh:
				case <-time.After(postTimeoutWait):
				}
				postTimeoutLogger(timedOutAt, result)
			}()
		}()
		return nil, errors.NewTimeoutError(fmt.Sprintf("request did not complete within requested timeout - %s", ctx.Err()), 0)
	}
}

// logPostTimeoutResult 在请求超时后，记录sender发送给接收方的结果的 panic 或错误。timedOutAt 是请求超时的时间
func logPostTimeoutResult(timedOutAt time.Time, r *result) {
	if r == nil {
		//	// we are using r == nil to indicate that the child goroutine never returned a result.
		//	metrics.RecordRequestPostTimeout(metrics.PostTimeoutSourceRestHandler, metrics.PostTimeoutHandlerPending)
		//	klog.Errorf("FinishRequest: post-timeout activity, waited for %s, child goroutine has not returned yet", time.Since(timedOutAt))
		return
	}
	//
	//var status string
	//switch {
	//case r.reason != nil:
	//	// a non empty reason inside a result object indicates that there was a panic.
	//	status = metrics.PostTimeoutHandlerPanic
	//case r.err != nil:
	//	status = metrics.PostTimeoutHandlerError
	//default:
	//	status = metrics.PostTimeoutHandlerOK
	//}
	//
	//metrics.RecordRequestPostTimeout(metrics.PostTimeoutSourceRestHandler, status)
	//err := fmt.Errorf("FinishRequest: post-timeout activity - time-elapsed: %s, panicked: %t, err: %v, panic-reason: %v",
	//	time.Since(timedOutAt), r.reason != nil, r.err, r.reason)
	//utilruntime.HandleError(err)
}
