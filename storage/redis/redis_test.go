package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/balramadan/gopherqueue"
	"github.com/balramadan/gopherqueue/storage/redis"
	redisv9 "github.com/redis/go-redis/v9"
)

func getTestRedisClient(t *testing.T) *redisv9.Client {
	client := redisv9.NewClient(&redisv9.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Skipping redis storage unit test; Redis not available at localhost:6379: %v", err)
	}

	return client
}

func TestRedisStore_Lifecycle(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.Close()

	ctx := context.Background()
	qName := "test_store_" + time.Now().Format("150405")
	store := redis.NewStore(client)

	// 1. Enqueue Instant Job
	job, err := gopherqueue.NewJob("job-store-1", qName, "email:send", map[string]any{"to": "user@example.com"})
	if err != nil {
		t.Fatalf("NewJob failed: %v", err)
	}

	err = store.Enqueue(ctx, job)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	pendingCount, err := store.GetPendingCount(ctx, qName)
	if err != nil || pendingCount != 1 {
		t.Errorf("expected pending count 1, got %d, err: %v", pendingCount, err)
	}

	// 2. Fetch Job
	fetchedJob, err := store.Fetch(ctx, qName, 2*time.Second)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	if fetchedJob.ID != "job-store-1" {
		t.Errorf("expected job ID 'job-store-1', got '%s'", fetchedJob.ID)
	}
	if fetchedJob.Status != gopherqueue.StatusProcessing {
		t.Errorf("expected status PROCESSING, got %s", fetchedJob.Status)
	}

	// 3. Reschedule (Rate Limiting)
	err = store.Reschedule(ctx, fetchedJob, 10*time.Second)
	if err != nil {
		t.Fatalf("Reschedule failed: %v", err)
	}

	delayedCount, _ := store.GetDelayedCount(ctx, qName)
	if delayedCount != 1 {
		t.Errorf("expected delayed count 1 after Reschedule, got %d", delayedCount)
	}

	processingCount, _ := store.GetProcessingCount(ctx, qName)
	if processingCount != 0 {
		t.Errorf("expected processing count 0 after Reschedule, got %d", processingCount)
	}

	// 4. Ack
	err = store.Ack(ctx, fetchedJob)
	if err != nil {
		t.Fatalf("Ack failed: %v", err)
	}
}
