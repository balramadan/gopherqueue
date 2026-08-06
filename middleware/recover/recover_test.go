package recover_test

import (
	"context"
	"strings"
	"testing"

	"github.com/balramadan/gopherqueue"
	mwrecover "github.com/balramadan/gopherqueue/middleware/recover"
)

func TestRecoverMiddleware_CapturesPanic(t *testing.T) {
	handler := gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
		panic("simulated nil pointer deference")
	})

	mw := mwrecover.New()
	pipeline := mw(handler)

	job, _ := gopherqueue.NewJob("job-panic-1", "default", "crash:task", nil)
	err := pipeline.ProcessJob(context.Background(), job)

	if err == nil {
		t.Fatalf("expected error from panic recovery, got nil")
	}

	if !strings.Contains(err.Error(), "panic recovered during job execution: simulated nil pointer deference") {
		t.Errorf("unexpected error message: %v", err)
	}
}
