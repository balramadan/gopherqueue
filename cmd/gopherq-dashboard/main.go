package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/balramadan/gopherqueue/storage/redis"
	"github.com/balramadan/gopherqueue/telemetry"
	"github.com/balramadan/gopherqueue/web"
	redisv9 "github.com/redis/go-redis/v9"
)

func main() {
	log.Println("Starting GopherQueue Standalone Web Dashboard Server...")

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	rdb := redisv9.NewClient(&redisv9.Options{Addr: redisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s: %v", redisAddr, err)
	}

	redisStore := redis.NewStore(rdb)
	sseBroadcaster := telemetry.NewSSEBroadcaster(redisStore, "default")

	mux := http.NewServeMux()

	// SSE Stream Endpoint
	mux.HandleFunc("/api/v1/stream", sseBroadcaster.StreamMetrics)

	// Live Metrics Snapshot Endpoint
	mux.HandleFunc("/api/v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		snapshot, err := telemetry.CollectMetrics(r.Context(), redisStore, "default")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(snapshot)
	})

	// Healthcheck
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Serve embedded Vue Web Dashboard static assets (Zero dependency)
	if distFS, err := web.DistFS(); err == nil {
		log.Println("Serving Embedded Vue Web Dashboard GUI...")
		fileServer := http.FileServer(http.FS(distFS))
		mux.Handle("/", fileServer)
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<h1>GopherQueue Dashboard API</h1><p>Connect SSE to <code>/api/v1/stream</code></p>"))
		})
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("GopherQueue Web Dashboard listening on port http://localhost:%s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Dashboard HTTP server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("Shutdown signal received. Shutting down Dashboard server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Dashboard server forced shutdown: %v", err)
	}

	log.Println("GopherQueue Web Dashboard stopped cleanly.")
}
