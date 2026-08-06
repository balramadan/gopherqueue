package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/balramadan/gopherqueue"
	gopherqlogger "github.com/balramadan/gopherqueue/middleware/logger"
	gopherqrecover "github.com/balramadan/gopherqueue/middleware/recover"
)

func main() {
	log.Println("--- GopherQueue Middleware Pipeline Example ---")

	slogger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 1. Initialize Server
	server := gopherqueue.NewServer(
		gopherqueue.WithConcurrency(2),
	)

	// 2. Attach Built-in Middlewares (Logger & Recover Guard)
	server.Use(gopherqlogger.New(slogger))
	server.Use(gopherqrecover.New())

	// 3. Attach Custom Audit Middleware
	server.Use(func(next gopherqueue.Handler) gopherqueue.Handler {
		return gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
			fmt.Printf(" [Audit MW] Inspecting Job %s before execution\n", job.ID)
			return next.ProcessJob(ctx, job)
		})
	})

	// 4. Register Handler
	server.Register("task:process", func(ctx context.Context, job *gopherqueue.Job) error {
		fmt.Printf(" [Handler] Processing task ID: %s\n", job.ID)
		return nil
	})

	log.Println("Middleware pipeline configured cleanly.")
}
