package distlimit

import (
	"context"
	"fmt"
	"time"

	"github.com/balramadan/distlimit"
	"github.com/balramadan/gopherqueue"
)

// KeyExtractor mengambil kunci rate limit dari metadata Job.
type KeyExtractor func(job *gopherqueue.Job) string

// DefaultKeyExtractor mengambil kunci berdasarkan tipe job.
func DefaultKeyExtractor(job *gopherqueue.Job) string {
	if job == nil {
		return ""
	}
	return "rate_limit:" + job.Type
}

// New membuat gopherqueue.MiddlewareFunc menggunakan instance *distlimit.Limiter.
func New(limiter *distlimit.Limiter, keyExtractor KeyExtractor) gopherqueue.MiddlewareFunc {
	if keyExtractor == nil {
		keyExtractor = DefaultKeyExtractor
	}

	return func(next gopherqueue.Handler) gopherqueue.Handler {
		return gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
			if job == nil || limiter == nil {
				return next.ProcessJob(ctx, job)
			}

			limitKey := keyExtractor(job)
			if limitKey == "" {
				return next.ProcessJob(ctx, job)
			}

			res, err := limiter.AllowKey(ctx, limitKey)
			if err != nil {
				return fmt.Errorf("gopherqueue/distlimit: rate limit evaluation error: %w", err)
			}

			if !res.Allowed {
				retryAfter := res.ResetIn
				if retryAfter <= 0 {
					retryAfter = 5 * time.Second
				}
				return &gopherqueue.ErrRateLimited{
					Key:        limitKey,
					RetryAfter: retryAfter,
				}
			}

			return next.ProcessJob(ctx, job)
		})
	}
}
