package distlimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/balramadan/distlimit"
	"github.com/balramadan/distlimit/algorithm/tokenbucket"
	"github.com/balramadan/distlimit/driver/memory"
	"github.com/balramadan/gopherqueue"
	mwdist "github.com/balramadan/gopherqueue/middleware/distlimit"
)

func TestDistlimitMiddleware_AllowedAndBlocked(t *testing.T) {
	memDriver := memory.New(5 * time.Minute)
	defer memDriver.Close(context.Background())

	limiter, err := distlimit.New(
		memDriver,
		distlimit.WithLimit(1), // Allow 1 request per minute
		distlimit.WithWindow(1*time.Minute),
		distlimit.WithAlgorithm(tokenbucket.New()),
	)
	if err != nil {
		t.Fatalf("failed to init distlimit: %v", err)
	}

	executedCount := 0
	handler := gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
		executedCount++
		return nil
	})

	mw := mwdist.New(limiter, func(job *gopherqueue.Job) string {
		return "job_limit_key"
	})
	pipeline := mw(handler)

	job1, _ := gopherqueue.NewJob("job-dist-1", "default", "email:send", nil)
	err1 := pipeline.ProcessJob(context.Background(), job1)
	if err1 != nil {
		t.Fatalf("first job should be allowed, got: %v", err1)
	}

	job2, _ := gopherqueue.NewJob("job-dist-2", "default", "email:send", nil)
	err2 := pipeline.ProcessJob(context.Background(), job2)
	if err2 == nil {
		t.Fatalf("second job should be rate limited, got nil")
	}

	rateErr, ok := gopherqueue.IsRateLimited(err2)
	if !ok {
		t.Fatalf("expected ErrRateLimited, got %T: %v", err2, err2)
	}

	if rateErr.Key != "job_limit_key" {
		t.Errorf("expected key 'job_limit_key', got '%s'", rateErr.Key)
	}
	if executedCount != 1 {
		t.Errorf("expected handler execution count 1, got %d", executedCount)
	}
}
