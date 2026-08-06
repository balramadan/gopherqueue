package gopherqueue_test

import (
	"context"
	"errors"
	"testing"

	"github.com/balramadan/gopherqueue"
)

func TestMiddlewareChain_ExecutionOrder(t *testing.T) {
	var executionOrder []string

	mw1 := func(next gopherqueue.Handler) gopherqueue.Handler {
		return gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
			executionOrder = append(executionOrder, "mw1_start")
			err := next.ProcessJob(ctx, job)
			executionOrder = append(executionOrder, "mw1_end")
			return err
		})
	}

	mw2 := func(next gopherqueue.Handler) gopherqueue.Handler {
		return gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
			executionOrder = append(executionOrder, "mw2_start")
			err := next.ProcessJob(ctx, job)
			executionOrder = append(executionOrder, "mw2_end")
			return err
		})
	}

	finalHandler := gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
		executionOrder = append(executionOrder, "handler")
		return nil
	})

	chain := gopherqueue.NewChain(mw1, mw2)
	pipeline := chain.Then(finalHandler)

	job, _ := gopherqueue.NewJob("job-mw-1", "default", "test:type", nil)
	err := pipeline.ProcessJob(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedOrder := []string{"mw1_start", "mw2_start", "handler", "mw2_end", "mw1_end"}
	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("expected length %d, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, v := range expectedOrder {
		if executionOrder[i] != v {
			t.Errorf("at index %d: expected %s, got %s", i, v, executionOrder[i])
		}
	}
}

func TestMiddlewareChain_Append(t *testing.T) {
	mw1 := func(next gopherqueue.Handler) gopherqueue.Handler {
		return gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
			return next.ProcessJob(ctx, job)
		})
	}
	mw2 := func(next gopherqueue.Handler) gopherqueue.Handler {
		return gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
			return errors.New("mw2 error")
		})
	}

	chain1 := gopherqueue.NewChain(mw1)
	chain2 := chain1.Append(mw2)

	job, _ := gopherqueue.NewJob("job-append-1", "default", "test:type", nil)
	err := chain2.Then(nil).ProcessJob(context.Background(), job)
	if err == nil || err.Error() != "mw2 error" {
		t.Errorf("expected 'mw2 error', got %v", err)
	}
}
