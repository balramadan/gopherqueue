package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/balramadan/gopherqueue"
	gopherqredis "github.com/balramadan/gopherqueue/storage/redis"
	"github.com/redis/go-redis/v9"
)

func main() {
	log.Println("--- GopherQueue Basic Embedded Mode Example ---")

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("Notice: Redis not available on localhost:6379: %v", err)
		log.Println("This example demonstrates code structure for embedded server and client.")
		return
	}

	redisStore := gopherqredis.NewStore(rdb)

	// 1. Initialize Producer Client
	client, err := gopherqueue.NewClient(
		gopherqueue.WithStorage(redisStore),
		gopherqueue.WithQueue("default"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Client: %v", err)
	}

	// 2. Initialize Worker Server
	server := gopherqueue.NewServer(
		gopherqueue.WithStorage(redisStore),
		gopherqueue.WithQueue("default"),
		gopherqueue.WithConcurrency(3),
	)

	// 3. Register Job Handler
	server.Register("email:welcome", func(ctx context.Context, job *gopherqueue.Job) error {
		email, _ := job.Payload["email"].(string)
		fmt.Printf(" [Worker] Sending welcome email to %s (Job ID: %s)\n", email, job.ID)
		return nil
	})

	// 4. Enqueue Jobs
	job1, _ := gopherqueue.NewJob("job-welcome-1", "default", "email:welcome", map[string]any{
		"email": "user1@example.com",
	})
	_ = client.Enqueue(context.Background(), job1)

	job2, _ := gopherqueue.NewJob("job-welcome-2", "default", "email:welcome", map[string]any{
		"email": "user2@example.com",
	})
	_ = client.EnqueueAt(context.Background(), job2, time.Now().Add(2*time.Second))

	// 5. Run Server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Running GopherQueue Worker Server for 5 seconds...")
	if err := server.Run(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Basic example finished successfully.")
}
