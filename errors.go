package gopherqueue

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrJobNotFound indicates that the requested job ID does not exist in storage.
	ErrJobNotFound = errors.New("gopherqueue: job not found")
	// ErrQueueEmpty indicates that no pending job is available for fetching.
	ErrQueueEmpty = errors.New("gopherqueue: queue is empty")
	// ErrInvalidJobData indicates missing required fields when creating or processing a job.
	ErrInvalidJobData = errors.New("gopherqueue: invalid job data or metadata")
	// ErrServerClosed indicates operations were attempted on a closed server.
	ErrServerClosed = errors.New("gopherqueue: server is closed")
)

// ErrRateLimited is a custom error returned by rate limiting middleware to signal
// that execution was rate limited and the job should be deferred to the delayed queue.
type ErrRateLimited struct {
	Key        string
	RetryAfter time.Duration
}

// Error returns a formatted error message.
func (e *ErrRateLimited) Error() string {
	return fmt.Sprintf("gopherqueue: rate limit exceeded for key '%s', retry after %v", e.Key, e.RetryAfter)
}

// IsRateLimited inspects an error chain and returns *ErrRateLimited if present.
func IsRateLimited(err error) (*ErrRateLimited, bool) {
	var rateErr *ErrRateLimited
	if errors.As(err, &rateErr) {
		return rateErr, true
	}
	return nil, false
}
