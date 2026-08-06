package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/balramadan/gopherqueue"
	"github.com/redis/go-redis/v9"
)

var _ gopherqueue.Storage = (*Store)(nil)

type Store struct {
	client *redis.Client
}

func NewStore(client *redis.Client) *Store {
	return &Store{client: client}
}

func keyPending(queue string) string    { return fmt.Sprintf("gopherq:queue:%s:pending", queue) }
func keyProcessing(queue string) string { return fmt.Sprintf("gopherq:queue:%s:processing", queue) }
func keyDelayed(queue string) string    { return fmt.Sprintf("gopherq:queue:%s:delayed", queue) }
func keyDLQ(queue string) string        { return fmt.Sprintf("gopherq:queue:%s:dlq", queue) }
func keyJobData(jobID string) string    { return fmt.Sprintf("gopherq:job:%s", jobID) }

func (s *Store) Enqueue(ctx context.Context, job *gopherqueue.Job) error {
	if job.Queue == "" {
		job.Queue = "default"
	}
	now := time.Now().Unix()
	job.UpdatedAt = now

	data, err := job.Marshal()
	if err != nil {
		return err
	}

	pipe := s.client.Pipeline()
	pipe.Set(ctx, keyJobData(job.ID), data, 7*24*time.Hour)

	if job.ExecuteAt > now {
		job.Status = gopherqueue.StatusDelayed
		pipe.ZAdd(ctx, keyDelayed(job.Queue), redis.Z{
			Score:  float64(job.ExecuteAt),
			Member: job.ID,
		})
	} else {
		job.Status = gopherqueue.StatusPending
		pipe.LPush(ctx, keyPending(job.Queue), job.ID)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) Fetch(ctx context.Context, queue string, timeout time.Duration) (*gopherqueue.Job, error) {
	// Atomic Fetch via BLMOVE (Pending -> Processing)
	jobID, err := s.client.BLMove(ctx, keyPending(queue), keyProcessing(queue), "RIGHT", "LEFT", timeout).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, gopherqueue.ErrQueueEmpty
		}
		return nil, fmt.Errorf("gopherqueue/redis: BLMOVE error: %w", err)
	}

	jobDataStr, err := s.client.Get(ctx, keyJobData(jobID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, gopherqueue.ErrJobNotFound
		}
		return nil, err
	}

	job, err := gopherqueue.UnmarshalJob(jobDataStr)
	if err != nil {
		return nil, err
	}

	job.Status = gopherqueue.StatusProcessing
	job.UpdatedAt = time.Now().Unix()

	updatedData, _ := job.Marshal()
	s.client.Set(ctx, keyJobData(job.ID), updatedData, 7*24*time.Hour)

	return job, nil
}

func (s *Store) Ack(ctx context.Context, job *gopherqueue.Job) error {
	now := time.Now().Unix()
	job.Status = gopherqueue.StatusCompleted
	job.UpdatedAt = now

	updatedData, err := job.Marshal()
	if err != nil {
		return err
	}

	pipe := s.client.Pipeline()
	pipe.LRem(ctx, keyProcessing(job.Queue), 1, job.ID)
	pipe.Set(ctx, keyJobData(job.ID), updatedData, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) Nack(ctx context.Context, job *gopherqueue.Job, execErr error) error {
	now := time.Now().Unix()
	job.CurrentRetry++
	if execErr != nil {
		job.LastError = execErr.Error()
	}

	pipe := s.client.Pipeline()
	pipe.LRem(ctx, keyProcessing(job.Queue), 1, job.ID)

	if job.CurrentRetry >= job.MaxRetries {
		job.Status = gopherqueue.StatusDeadLetter
		job.UpdatedAt = now

		data, _ := job.Marshal()
		pipe.Set(ctx, keyJobData(job.ID), data, 30*24*time.Hour)
		pipe.RPush(ctx, keyDLQ(job.Queue), job.ID)
	} else {
		job.Status = gopherqueue.StatusRetrying
		delay := job.CalculateNextRetryDelay()
		job.ExecuteAt = now + int64(delay.Seconds())
		job.UpdatedAt = now

		data, _ := job.Marshal()
		pipe.Set(ctx, keyJobData(job.ID), data, 7*24*time.Hour)
		pipe.ZAdd(ctx, keyDelayed(job.Queue), redis.Z{
			Score:  float64(job.ExecuteAt),
			Member: job.ID,
		})
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Reschedule memindahkan job kembali ke Delayed ZSet tanpa memotong kuota retry (Khusus Rate Limiting).
func (s *Store) Reschedule(ctx context.Context, job *gopherqueue.Job, delay time.Duration) error {
	now := time.Now().Unix()
	job.Status = gopherqueue.StatusDelayed
	job.ExecuteAt = now + int64(delay.Seconds())
	job.UpdatedAt = now
	job.LastError = fmt.Sprintf("deferred due to rate limit, retry after %v", delay)

	data, err := job.Marshal()
	if err != nil {
		return err
	}

	pipe := s.client.Pipeline()
	// 1. Hapus dari list processing
	pipe.LRem(ctx, keyProcessing(job.Queue), 1, job.ID)

	// 2. Simpan metadata baru
	pipe.Set(ctx, keyJobData(job.ID), data, 7*24*time.Hour)

	// 3. Masukkan ke delayed ZSet dengan score timestamp baru
	pipe.ZAdd(ctx, keyDelayed(job.Queue), redis.Z{
		Score:  float64(job.ExecuteAt),
		Member: job.ID,
	})

	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) GetPendingCount(ctx context.Context, queue string) (int64, error) {
	return s.client.LLen(ctx, keyPending(queue)).Result()
}

func (s *Store) GetProcessingCount(ctx context.Context, queue string) (int64, error) {
	return s.client.LLen(ctx, keyProcessing(queue)).Result()
}

func (s *Store) GetDelayedCount(ctx context.Context, queue string) (int64, error) {
	return s.client.ZCard(ctx, keyDelayed(queue)).Result()
}

func (s *Store) GetDLQCount(ctx context.Context, queue string) (int64, error) {
	return s.client.LLen(ctx, keyDLQ(queue)).Result()
}
