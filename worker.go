package gopherqueue

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// JobHandler represents a handler function signature for processing a job.
type JobHandler func(ctx context.Context, job *Job) error

// JobExecutor defines an interface for executing a job.
type JobExecutor interface {
	Execute(ctx context.Context, job *Job) error
}

// HandlerRegistry maintains a thread-safe mapping of job types to JobHandlers.
type HandlerRegistry struct {
	mu       sync.RWMutex
	handlers map[string]JobHandler
}

// NewHandlerRegistry initializes a new HandlerRegistry instance.
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string]JobHandler),
	}
}

// Register registers a handler function for a given job type.
func (r *HandlerRegistry) Register(jobType string, handler JobHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[jobType] = handler
}

// Execute looks up and invokes the registered handler for the job type.
func (r *HandlerRegistry) Execute(ctx context.Context, job *Job) error {
	r.mu.RLock()
	handler, exist := r.handlers[job.Type]
	r.mu.RUnlock()

	if !exist {
		return fmt.Errorf("gopherqueue: no handler registered for job type '%s'", job.Type)
	}

	return handler(ctx, job)
}

// WorkerPool manages worker goroutines that fetch and process jobs concurrently.
type WorkerPool struct {
	queueName    string
	storage      Storage
	executor     JobExecutor
	numWorkers   int
	pollInterval time.Duration
	wg           sync.WaitGroup
	quit         chan struct{}
	stopOnce     sync.Once
}

// NewWorkerPool initializes a new WorkerPool instance.
func NewWorkerPool(queueName string, numWorkers int, pollInterval time.Duration, store Storage, exec JobExecutor) *WorkerPool {
	if numWorkers <= 0 {
		numWorkers = 5
	}
	if queueName == "" {
		queueName = "default"
	}
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}

	return &WorkerPool{
		queueName:    queueName,
		storage:      store,
		executor:     exec,
		numWorkers:   numWorkers,
		pollInterval: pollInterval,
		quit:         make(chan struct{}),
	}
}

// Start launches worker goroutines.
func (wp *WorkerPool) Start(ctx context.Context) {
	log.Printf("[WorkerPool] Starting %d workers for queue '%s'...", wp.numWorkers, wp.queueName)
	for i := 1; i <= wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

func (wp *WorkerPool) worker(ctx context.Context, workerID int) {
	defer wp.wg.Done()
	log.Printf("[Worker %d] Ready and listening for jobs on queue '%s'...", workerID, wp.queueName)

	for {
		select {
		case <-wp.quit:
			log.Printf("[Worker %d] Stopping...", workerID)
			return
		case <-ctx.Done():
			log.Printf("[Worker %d] Context cancelled. Stopping...", workerID)
			return
		default:
			job, err := wp.storage.Fetch(ctx, wp.queueName, wp.pollInterval)
			if err != nil {
				if errors.Is(err, ErrQueueEmpty) {
					continue
				}
				if errors.Is(err, context.Canceled) {
					return
				}

				log.Printf("[Worker %d] Fetch error: %v", workerID, err)
				time.Sleep(1 * time.Second)
				continue
			}

			wp.processJob(ctx, workerID, job)
		}
	}
}

func (wp *WorkerPool) processJob(parentCtx context.Context, workerID int, job *Job) {
	timeout := time.Duration(job.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	jobCtx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	log.Printf("[Worker %d] Processing Job ID: %s (Type: %s)", workerID, job.ID, job.Type)

	var execErr error

	defer func() {
		if r := recover(); r != nil {
			execErr = fmt.Errorf("panic recovered during execution: %v", r)
			log.Printf("[Worker %d] CRITICAL: Job %s panicked: %v", workerID, job.ID, r)
		}

		ackCtx, ackCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer ackCancel()

		if execErr == nil {
			if ackErr := wp.storage.Ack(ackCtx, job); ackErr != nil {
				log.Printf("[Worker %d] Failed to ACK Job %s: %v", workerID, job.ID, ackErr)
			} else {
				log.Printf("[Worker %d] Successfully COMPLETED Job ID: %s", workerID, job.ID)
			}
		} else if rateErr, ok := IsRateLimited(execErr); ok {
			if reschedErr := wp.storage.Reschedule(ackCtx, job, rateErr.RetryAfter); reschedErr != nil {
				log.Printf("[Worker %d] Failed to Reschedule Rate Limited Job %s: %v", workerID, job.ID, reschedErr)
			} else {
				log.Printf("[Worker %d] Job ID: %s RATE_LIMITED (Key: %s, Deferred for %v)", workerID, job.ID, rateErr.Key, rateErr.RetryAfter)
			}
		} else {
			if nackErr := wp.storage.Nack(ackCtx, job, execErr); nackErr != nil {
				log.Printf("[Worker %d] Failed to NACK Job %s: %v", workerID, job.ID, nackErr)
			} else {
				log.Printf("[Worker %d] Job ID: %s FAILED/RETRYING (Error: %v)", workerID, job.ID, execErr)
			}
		}
	}()

	execErr = wp.executor.Execute(jobCtx, job)
}

// Stop signals all workers to stop and waits for them to finish.
func (wp *WorkerPool) Stop() {
	wp.stopOnce.Do(func() {
		close(wp.quit)
	})
	wp.wg.Wait()
	log.Println("[WorkerPool] All workers stopped cleanly.")
}
