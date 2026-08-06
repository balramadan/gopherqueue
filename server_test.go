package gopherqueue_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/balramadan/gopherqueue"
)

type mockRootStorage struct {
	mu           sync.Mutex
	jobs         []*gopherqueue.Job
	ackedJobs    []string
	deferredJobs []string
}

func (m *mockRootStorage) Enqueue(ctx context.Context, job *gopherqueue.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs = append(m.jobs, job)
	return nil
}

func (m *mockRootStorage) Fetch(ctx context.Context, queue string, timeout time.Duration) (*gopherqueue.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.jobs) == 0 {
		time.Sleep(10 * time.Millisecond)
		return nil, gopherqueue.ErrQueueEmpty
	}
	job := m.jobs[0]
	m.jobs = m.jobs[1:]
	return job, nil
}

func (m *mockRootStorage) Ack(ctx context.Context, job *gopherqueue.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ackedJobs = append(m.ackedJobs, job.ID)
	return nil
}

func (m *mockRootStorage) Nack(ctx context.Context, job *gopherqueue.Job, execErr error) error {
	return nil
}

func (m *mockRootStorage) Reschedule(ctx context.Context, job *gopherqueue.Job, delay time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deferredJobs = append(m.deferredJobs, job.ID)
	return nil
}

func (m *mockRootStorage) GetPendingCount(ctx context.Context, queue string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return int64(len(m.jobs)), nil
}

func (m *mockRootStorage) GetProcessingCount(ctx context.Context, queue string) (int64, error) {
	return 0, nil
}

func (m *mockRootStorage) GetDelayedCount(ctx context.Context, queue string) (int64, error) {
	return 0, nil
}

func (m *mockRootStorage) GetDLQCount(ctx context.Context, queue string) (int64, error) {
	return 0, nil
}

func TestServer_RegisterAndMiddlewareExecution(t *testing.T) {
	job, _ := gopherqueue.NewJob("job-srv-1", "default", "email:send", nil)
	store := &mockRootStorage{jobs: []*gopherqueue.Job{job}}

	server := gopherqueue.NewServer(
		gopherqueue.WithStorage(store),
		gopherqueue.WithConcurrency(1),
		gopherqueue.WithPollInterval(20*time.Millisecond),
	)

	mwExecuted := false
	server.Use(func(next gopherqueue.Handler) gopherqueue.Handler {
		return gopherqueue.HandlerFunc(func(ctx context.Context, j *gopherqueue.Job) error {
			mwExecuted = true
			return next.ProcessJob(ctx, j)
		})
	})

	handlerExecuted := false
	server.Register("email:send", func(ctx context.Context, j *gopherqueue.Job) error {
		handlerExecuted = true
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_ = server.Run(ctx)

	if !mwExecuted {
		t.Errorf("expected middleware to be executed")
	}
	if !handlerExecuted {
		t.Errorf("expected handler to be executed")
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.ackedJobs) != 1 || store.ackedJobs[0] != "job-srv-1" {
		t.Errorf("expected job-srv-1 to be acked, got %v", store.ackedJobs)
	}
}
