package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/balramadan/gopherqueue"
	gopherqredis "github.com/balramadan/gopherqueue/storage/redis"
	"github.com/redis/go-redis/v9"
)

func main() {
	log.Println("Starting GopherQueue Standalone Server Daemon...")

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		queueName = "default"
	}

	concurrencyStr := os.Getenv("CONCURRENCY")
	concurrency := 10
	if concurrencyStr != "" {
		if c, err := strconv.Atoi(concurrencyStr); err == nil && c > 0 {
			concurrency = c
		}
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctxCheck, cancelCheck := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCheck()
	if err := rdb.Ping(ctxCheck).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s: %v", redisAddr, err)
	}

	redisStore := gopherqredis.NewStore(rdb)

	server := gopherqueue.NewServer(
		gopherqueue.WithStorage(redisStore),
		gopherqueue.WithQueue(queueName),
		gopherqueue.WithConcurrency(concurrency),
	)

	// Register sample handlers
	server.Register("email:send", func(ctx context.Context, job *gopherqueue.Job) error {
		log.Printf("[Daemon] Executing email:send job ID: %s", job.ID)
		return nil
	})

	server.Register("report:generate", func(ctx context.Context, job *gopherqueue.Job) error {
		log.Printf("[Daemon] Executing report:generate job ID: %s", job.ID)
		return nil
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("GopherQueue Standalone Daemon running on queue '%s' with %d workers...", queueName, concurrency)
	if err := server.Run(ctx); err != nil {
		log.Fatalf("Server daemon error: %v", err)
	}

	log.Println("GopherQueue Standalone Daemon stopped cleanly.")
}
