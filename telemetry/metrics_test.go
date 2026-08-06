package telemetry_test

import (
	"context"
	"testing"
	"time"

	"github.com/balramadan/gopherqueue"
	"github.com/balramadan/gopherqueue/telemetry"
)

type mockTelemetryStorage struct {
	pending    int64
	processing int64
	delayed    int64
	dlq        int64
}

func (m *mockTelemetryStorage) Enqueue(ctx context.Context, job *gopherqueue.Job) error { return nil }
func (m *mockTelemetryStorage) Fetch(ctx context.Context, queue string, timeout time.Duration) (*gopherqueue.Job, error) {
	return nil, gopherqueue.ErrQueueEmpty
}
func (m *mockTelemetryStorage) Ack(ctx context.Context, job *gopherqueue.Job) error              { return nil }
func (m *mockTelemetryStorage) Nack(ctx context.Context, job *gopherqueue.Job, execErr error) error { return nil }
func (m *mockTelemetryStorage) Reschedule(ctx context.Context, job *gopherqueue.Job, delay time.Duration) error {
	return nil
}
func (m *mockTelemetryStorage) GetPendingCount(ctx context.Context, queue string) (int64, error) {
	return m.pending, nil
}
func (m *mockTelemetryStorage) GetProcessingCount(ctx context.Context, queue string) (int64, error) {
	return m.processing, nil
}
func (m *mockTelemetryStorage) GetDelayedCount(ctx context.Context, queue string) (int64, error) {
	return m.delayed, nil
}
func (m *mockTelemetryStorage) GetDLQCount(ctx context.Context, queue string) (int64, error) {
	return m.dlq, nil
}

func TestCollectMetrics(t *testing.T) {
	store := &mockTelemetryStorage{
		pending:    10,
		processing: 2,
		delayed:    5,
		dlq:        1,
	}

	snapshot, err := telemetry.CollectMetrics(context.Background(), store, "default")
	if err != nil {
		t.Fatalf("CollectMetrics failed: %v", err)
	}

	if snapshot.Pending != 10 || snapshot.Processing != 2 || snapshot.Delayed != 5 || snapshot.DLQ != 1 {
		t.Errorf("unexpected snapshot metrics: %+v", snapshot)
	}
}
