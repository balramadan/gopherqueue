package recover

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/balramadan/gopherqueue"
)

// New membuat middleware penangkap panic agar worker goroutine tidak crash.
func New() gopherqueue.MiddlewareFunc {
	return func(next gopherqueue.Handler) gopherqueue.Handler {
		return gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) (err error) {
			defer func() {
				if r := recover(); r != nil {
					stack := debug.Stack()
					err = fmt.Errorf("panic recovered during job execution: %v\nstacktrace:\n%s", r, string(stack))
				}
			}()

			return next.ProcessJob(ctx, job)
		})
	}
}
