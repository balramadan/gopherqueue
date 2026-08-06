package gopherqueue_test

import (
	"context"
	"fmt"
	"time"

	"github.com/balramadan/gopherqueue"
)

// ExampleNewServer demonstrates how to initialize and run a GopherQueue worker server.
func ExampleNewServer() {
	store := &mockRootStorage{}

	server := gopherqueue.NewServer(
		gopherqueue.WithStorage(store),
		gopherqueue.WithQueue("default"),
		gopherqueue.WithConcurrency(2),
	)

	server.Register("email:send", func(ctx context.Context, job *gopherqueue.Job) error {
		fmt.Printf("Executing job %s of type %s\n", job.ID, job.Type)
		return nil
	})

	// Server is ready to run with server.Run(ctx)
	fmt.Println("Server initialized successfully")
	// Output: Server initialized successfully
}

// ExampleNewClient demonstrates how to enqueue immediate and delayed jobs.
func ExampleNewClient() {
	store := &mockRootStorage{}

	client, err := gopherqueue.NewClient(
		gopherqueue.WithStorage(store),
		gopherqueue.WithQueue("default"),
	)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	job, _ := gopherqueue.NewJob("job-101", "default", "email:send", map[string]any{
		"to": "user@example.com",
	})

	_ = client.Enqueue(context.Background(), job)
	_ = client.EnqueueAt(context.Background(), job, time.Now().Add(5*time.Minute))

	fmt.Println("Job enqueued successfully")
	// Output: Job enqueued successfully
}
