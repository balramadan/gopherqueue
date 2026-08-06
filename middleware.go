package gopherqueue

import "context"

// Handler defines the interface for executing a Job.
type Handler interface {
	ProcessJob(ctx context.Context, job *Job) error
}

// HandlerFunc is an adapter type allowing ordinary functions to act as a Handler.
type HandlerFunc func(ctx context.Context, job *Job) error

// ProcessJob calls the underlying HandlerFunc.
func (f HandlerFunc) ProcessJob(ctx context.Context, job *Job) error {
	return f(ctx, job)
}

// MiddlewareFunc defines a middleware function wrapper around a Handler.
type MiddlewareFunc func(next Handler) Handler

// Chain manages an ordered slice of MiddlewareFunc and compiles them into a single Handler.
type Chain struct {
	middlewares []MiddlewareFunc
}

// NewChain creates a new Chain instance with the provided middlewares.
func NewChain(middlewares ...MiddlewareFunc) Chain {
	return Chain{middlewares: append([]MiddlewareFunc(nil), middlewares...)}
}

// Then links the middleware chain to a final Job Handler.
func (c Chain) Then(final Handler) Handler {
	if final == nil {
		final = HandlerFunc(func(ctx context.Context, job *Job) error {
			return nil
		})
	}

	for i := len(c.middlewares) - 1; i >= 0; i-- {
		final = c.middlewares[i](final)
	}

	return final
}

// Append creates a new Chain by appending additional middlewares to the existing chain.
func (c Chain) Append(middlewares ...MiddlewareFunc) Chain {
	newMw := make([]MiddlewareFunc, 0, len(c.middlewares)+len(middlewares))
	newMw = append(newMw, c.middlewares...)
	newMw = append(newMw, middlewares...)
	return Chain{middlewares: newMw}
}
