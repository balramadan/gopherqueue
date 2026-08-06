package telemetry

import (
	"context"
	"time"

	"github.com/balramadan/gopherqueue"
)

type MetricsSnapshot struct {
	Pending    int64 `json:"pending"`
	Processing int64 `json:"processing"`
	Delayed    int64 `json:"delayed"`
	DLQ        int64 `json:"dlq"`
	Timestamp  int64 `json:"timestamp"`
}

func CollectMetrics(ctx context.Context, store gopherqueue.Storage, queueName string) (*MetricsSnapshot, error) {
	if store == nil {
		return nil, gopherqueue.ErrInvalidJobData
	}
	if queueName == "" {
		queueName = "default"
	}

	pending, _ := store.GetPendingCount(ctx, queueName)
	processing, _ := store.GetProcessingCount(ctx, queueName)
	delayed, _ := store.GetDelayedCount(ctx, queueName)
	dlq, _ := store.GetDLQCount(ctx, queueName)

	return &MetricsSnapshot{
		Pending:    pending,
		Processing: processing,
		Delayed:    delayed,
		DLQ:        dlq,
		Timestamp:  time.Now().Unix(),
	}, nil
}
