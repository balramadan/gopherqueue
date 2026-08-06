package gopherqueue_test

import (
	"context"
	"testing"
	"time"

	"github.com/balramadan/gopherqueue"
)

func TestClient_Enqueue(t *testing.T) {
	store := &mockRootStorage{}
	client, err := gopherqueue.NewClient(
		gopherqueue.WithStorage(store),
		gopherqueue.WithQueue("custom_queue"),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	job, _ := gopherqueue.NewJob("client-job-1", "", "sms:send", nil)
	err = client.Enqueue(context.Background(), job)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	if job.Queue != "custom_queue" {
		t.Errorf("expected queue 'custom_queue', got '%s'", job.Queue)
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.jobs) != 1 || store.jobs[0].ID != "client-job-1" {
		t.Errorf("expected enqueued job ID 'client-job-1'")
	}
}

func TestClient_EnqueueAt(t *testing.T) {
	store := &mockRootStorage{}
	client, err := gopherqueue.NewClient(gopherqueue.WithStorage(store))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	job, _ := gopherqueue.NewJob("client-job-2", "default", "sms:send", nil)
	futureTime := time.Now().Add(10 * time.Minute)
	err = client.EnqueueAt(context.Background(), job, futureTime)
	if err != nil {
		t.Fatalf("EnqueueAt failed: %v", err)
	}

	if job.ExecuteAt != futureTime.Unix() {
		t.Errorf("expected ExecuteAt %d, got %d", futureTime.Unix(), job.ExecuteAt)
	}
}
