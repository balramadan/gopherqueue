package logger_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/balramadan/gopherqueue"
	mwlogger "github.com/balramadan/gopherqueue/middleware/logger"
)

func TestLoggerMiddleware_Success(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
		return nil
	})

	mw := mwlogger.New(logger)
	pipeline := mw(handler)

	job, _ := gopherqueue.NewJob("job-log-1", "default", "email:send", nil)
	err := pipeline.ProcessJob(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "job processing started") || !strings.Contains(out, "job processing completed") {
		t.Errorf("expected start and completion logs, got: %s", out)
	}
}

func TestLoggerMiddleware_RateLimited(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
		return &gopherqueue.ErrRateLimited{Key: "limit:user_1", RetryAfter: 5 * time.Second}
	})

	mw := mwlogger.New(logger)
	pipeline := mw(handler)

	job, _ := gopherqueue.NewJob("job-log-2", "default", "email:send", nil)
	err := pipeline.ProcessJob(context.Background(), job)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	out := buf.String()
	if !strings.Contains(out, "job rate limited and deferred") {
		t.Errorf("expected rate limited log warning, got: %s", out)
	}
}

func TestLoggerMiddleware_Failed(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
		return errors.New("db connection timeout")
	})

	mw := mwlogger.New(logger)
	pipeline := mw(handler)

	job, _ := gopherqueue.NewJob("job-log-3", "default", "email:send", nil)
	err := pipeline.ProcessJob(context.Background(), job)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	out := buf.String()
	if !strings.Contains(out, "job processing failed") {
		t.Errorf("expected failed log, got: %s", out)
	}
}
