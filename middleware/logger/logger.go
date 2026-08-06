package logger

import (
	"context"
	"log/slog"
	"time"

	"github.com/balramadan/gopherqueue"
)

// New membuat middleware structured logging menggunakan log/slog.
func New(logger *slog.Logger) gopherqueue.MiddlewareFunc {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next gopherqueue.Handler) gopherqueue.Handler {
		return gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
			start := time.Now()

			logger.InfoContext(ctx, "job processing started",
				slog.String("job_id", job.ID),
				slog.String("queue", job.Queue),
				slog.String("type", job.Type),
				slog.Int("attempt", job.CurrentRetry+1),
			)

			err := next.ProcessJob(ctx, job)
			duration := time.Since(start)

			if err != nil {
				if rateErr, ok := gopherqueue.IsRateLimited(err); ok {
					logger.WarnContext(ctx, "job rate limited and deferred",
						slog.String("job_id", job.ID),
						slog.String("type", job.Type),
						slog.String("rate_key", rateErr.Key),
						slog.Duration("retry_after", rateErr.RetryAfter),
						slog.Duration("duration", duration),
					)
				} else {
					logger.ErrorContext(ctx, "job processing failed",
						slog.String("job_id", job.ID),
						slog.String("type", job.Type),
						slog.String("error", err.Error()),
						slog.Duration("duration", duration),
					)
				}
				return err
			}

			logger.InfoContext(ctx, "job processing completed",
				slog.String("job_id", job.ID),
				slog.String("type", job.Type),
				slog.Duration("duration", duration),
			)

			return nil
		})
	}
}
