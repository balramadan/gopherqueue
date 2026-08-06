package gopherqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// JobStatus represents the state of a job in the queue lifecycle.
type JobStatus string

const (
	// StatusEnqueued indicates the job has been submitted to the queue.
	StatusEnqueued JobStatus = "ENQUEUED"
	// StatusPending indicates the job is ready to be fetched by a worker.
	StatusPending JobStatus = "PENDING"
	// StatusDelayed indicates the job is scheduled to run at a future timestamp.
	StatusDelayed JobStatus = "DELAYED"
	// StatusProcessing indicates a worker is actively processing the job.
	StatusProcessing JobStatus = "PROCESSING"
	// StatusCompleted indicates the job finished successfully and was ACKed.
	StatusCompleted JobStatus = "COMPLETED"
	// StatusFailed indicates job execution failed.
	StatusFailed JobStatus = "FAILED"
	// StatusRetrying indicates the job failed and is scheduled for a retry.
	StatusRetrying JobStatus = "RETRYING"
	// StatusDeadLetter indicates the job exceeded its maximum retry budget and was moved to DLQ.
	StatusDeadLetter JobStatus = "DEAD_LETTER"
	// StatusRateLimited indicates job execution was rate-limited and deferred without incrementing retry count.
	StatusRateLimited JobStatus = "RATE_LIMITED"
)

// Job represents a unit of work submitted to GopherQueue.
type Job struct {
	ID             string         `json:"id"`
	Queue          string         `json:"queue"`
	Type           string         `json:"type"`
	Payload        map[string]any `json:"payload"`
	Status         JobStatus      `json:"status"`
	MaxRetries     int            `json:"max_retries"`
	CurrentRetry   int            `json:"current_retry"`
	TimeoutSeconds int            `json:"timeout_seconds"`
	ExecuteAt      int64          `json:"execute_at"`
	CreatedAt      int64          `json:"created_at"`
	UpdatedAt      int64          `json:"updated_at"`
	LastError      string         `json:"last_error,omitempty"`
}

// Storage defines the pluggable storage driver interface required by GopherQueue.
type Storage interface {
	Enqueue(ctx context.Context, job *Job) error
	Fetch(ctx context.Context, queue string, timeout time.Duration) (*Job, error)
	Ack(ctx context.Context, job *Job) error
	Nack(ctx context.Context, job *Job, execErr error) error
	Reschedule(ctx context.Context, job *Job, delay time.Duration) error
	GetPendingCount(ctx context.Context, queue string) (int64, error)
	GetProcessingCount(ctx context.Context, queue string) (int64, error)
	GetDelayedCount(ctx context.Context, queue string) (int64, error)
	GetDLQCount(ctx context.Context, queue string) (int64, error)
}

// NewJob creates a new Job instance with default configuration.
func NewJob(id, queue, jobType string, payload map[string]any) (*Job, error) {
	if id == "" || jobType == "" {
		return nil, ErrInvalidJobData
	}

	now := time.Now().Unix()
	return &Job{
		ID:             id,
		Queue:          queue,
		Type:           jobType,
		Payload:        payload,
		Status:         StatusEnqueued,
		MaxRetries:     3,
		CurrentRetry:   0,
		TimeoutSeconds: 30,
		ExecuteAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// Marshal serializes the Job into a JSON string.
func (j *Job) Marshal() (string, error) {
	bytes, err := json.Marshal(j)
	if err != nil {
		return "", fmt.Errorf("gopherqueue: failed to marshal job: %w", err)
	}

	return string(bytes), nil
}

// UnmarshalJob deserializes a JSON string into a Job instance.
func UnmarshalJob(data string) (*Job, error) {
	var j Job
	if err := json.Unmarshal([]byte(data), &j); err != nil {
		return nil, fmt.Errorf("gopherqueue: failed to unmarshal job: %w", err)
	}

	return &j, nil
}

// CalculateNextRetryDelay calculates an exponential backoff delay based on CurrentRetry.
func (j *Job) CalculateNextRetryDelay() time.Duration {
	baseInterval := 2.0
	backoff := baseInterval * math.Pow(2, float64(j.CurrentRetry))

	if backoff > 3600 {
		backoff = 3600
	}

	return time.Duration(backoff) * time.Second
}
