package gopherqueue

import (
	"context"
	"errors"
	"time"
)

// Client provides a producer API for submitting jobs to GopherQueue.
type Client struct {
	cfg     *Config
	storage Storage
}

// NewClient initializes a new Client instance with functional options.
func NewClient(opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.Storage == nil {
		return nil, errors.New("gopherqueue: storage engine is required")
	}

	return &Client{
		cfg:     cfg,
		storage: cfg.Storage,
	}, nil
}

// Enqueue submits a job to the queue for immediate processing.
func (c *Client) Enqueue(ctx context.Context, job *Job) error {
	if job == nil {
		return ErrInvalidJobData
	}
	if job.Queue == "" {
		job.Queue = c.cfg.QueueName
	}
	return c.storage.Enqueue(ctx, job)
}

// EnqueueAt schedules a job to be executed at a specific future timestamp.
func (c *Client) EnqueueAt(ctx context.Context, job *Job, executeAt time.Time) error {
	if job == nil {
		return ErrInvalidJobData
	}
	job.ExecuteAt = executeAt.Unix()
	if job.Queue == "" {
		job.Queue = c.cfg.QueueName
	}
	return c.storage.Enqueue(ctx, job)
}
