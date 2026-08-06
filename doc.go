/*
Package gopherqueue is an open-source, high-performance distributed task processing library and background job engine for Go.

GopherQueue supports Dual-Mode deployment:
1. Embedded Mode: Imported directly as a Go package (`go get github.com/balramadan/gopherqueue`) inside monolith or microservice Go applications.
2. Standalone Mode: Executed as a standalone worker daemon (`cmd/gopherq-server`) and web dashboard (`cmd/gopherq-dashboard`).

Key Features:
  - At-Least-Once Atomic Delivery: Powered by Redis BLMOVE operations to prevent job loss during worker crashes.
  - First-Class Rate Limiting Synergy: Native integration with distlimit (`github.com/balramadan/distlimit`). Rate-limited jobs are gracefully deferred to the delayed queue without consuming retry budgets.
  - HTTP-Style Middleware Pipeline: Expressive Chi-like middleware chaining (`Use(...)`) supporting structured logging (`slog`), panic recovery, and rate-limiting.
  - Zero-Dependency Telemetry: Server-Sent Events (SSE) broadcaster delivering real-time queue metrics to a Vue.js web dashboard.

Basic Usage:

	package main

	import (
		"context"
		"fmt"
		"log"

		"github.com/balramadan/gopherqueue"
		gopherqredis "github.com/balramadan/gopherqueue/storage/redis"
		"github.com/redis/go-redis/v9"
	)

	func main() {
		rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
		store := gopherqredis.NewStore(rdb)

		server := gopherqueue.NewServer(
			gopherqueue.WithStorage(store),
			gopherqueue.WithQueue("default"),
			gopherqueue.WithConcurrency(5),
		)

		server.Register("email:send", func(ctx context.Context, job *gopherqueue.Job) error {
			fmt.Printf("Processing job %s\n", job.ID)
			return nil
		})

		if err := server.Run(context.Background()); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
*/
package gopherqueue
