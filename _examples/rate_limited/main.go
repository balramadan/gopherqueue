package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/balramadan/distlimit"
	"github.com/balramadan/distlimit/algorithm/tokenbucket"
	distlimitredis "github.com/balramadan/distlimit/driver/redis"
	"github.com/balramadan/gopherqueue"
	gopherqdist "github.com/balramadan/gopherqueue/middleware/distlimit"
	gopherqredis "github.com/balramadan/gopherqueue/storage/redis"
	"github.com/redis/go-redis/v9"
)

func main() {
	log.Println("--- GopherQueue + distlimit Rate Limiting Example ---")

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("Notice: Redis not available on localhost:6379: %v", err)
		log.Println("This example demonstrates distlimit rate-limiting middleware integration.")
		return
	}

	// 1. Initialize distlimit Rate Limiter (Allow 2 requests per minute)
	limiterDriver := distlimitredis.New(rdb)
	limiter, err := distlimit.New(
		limiterDriver,
		distlimit.WithLimit(2),
		distlimit.WithWindow(1*time.Minute),
		distlimit.WithAlgorithm(tokenbucket.New()),
	)
	if err != nil {
		log.Fatalf("Failed to initialize distlimit: %v", err)
	}

	redisStore := gopherqredis.NewStore(rdb)

	// 2. Initialize GopherQueue Server
	server := gopherqueue.NewServer(
		gopherqueue.WithStorage(redisStore),
		gopherqueue.WithQueue("api_queue"),
		gopherqueue.WithConcurrency(2),
	)

	// 3. Attach distlimit Execution Guard Middleware
	server.Use(gopherqdist.New(limiter, func(job *gopherqueue.Job) string {
		// Rate limit per tenant or job type
		return "rate_limit:" + job.Type
	}))

	// 4. Register Job Handler
	server.Register("api:call", func(ctx context.Context, job *gopherqueue.Job) error {
		fmt.Printf(" [Worker] Executing rate-sensitive API call for job ID: %s\n", job.ID)
		return nil
	})

	log.Println("Server initialized with distlimit rate limiting guard.")
}
