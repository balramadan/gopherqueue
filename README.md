# 🚀 GopherQueue v0.1.0

[![Go Reference](https://pkg.go.dev/badge/github.com/balramadan/gopherqueue.svg)](https://pkg.go.dev/github.com/balramadan/gopherqueue)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/badge/Release-v0.1.0-blue.svg)](https://github.com/balramadan/gopherqueue/releases/tag/v0.1.0)

**`GopherQueue v0.1.0`** is an open-source, high-performance distributed task processing library and background job engine for Go. Engineered for both **embedded Go applications** and **standalone microservice daemons**, GopherQueue features **at-least-once atomic Redis delivery (`BLMOVE`)**, **first-class rate-limiting synergy with [`distlimit`](https://github.com/balramadan/distlimit)**, an **HTTP-style middleware pipeline**, and **zero-dependency Server-Sent Events (SSE) telemetry** paired with a Vue.js web dashboard.

---

## Table of Contents

- [⚡ Unique Value Propositions (UVP)](#-unique-value-propositions-uvp)
- [📊 Competitive Matrix](#-competitive-matrix)
- [📦 Installation](#-installation)
- [⚡ Quickstart](#-quickstart)
  - [1. Embedded Producer Client & Worker Server](#1-embedded-producer-client--worker-server)
  - [2. Integration with distlimit Rate Limiter](#2-integration-with-distlimit-rate-limiter)
- [🛠️ Middleware Ecosystem](#️-middleware-ecosystem)
- [🖥️ Zero-Dependency Web Dashboard GUI](#️-zero-dependency-web-dashboard-gui)
  - [Option 1: Standalone Single-Binary Execution (Zero Setup)](#option-1-standalone-single-binary-execution-zero-setup)
  - [Option 2: Embedded Route in your Go Application](#option-2-embedded-route-in-your-go-application)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)

---

## ⚡ Unique Value Propositions (UVP)

- **🔄 Dual-Mode Deployment:** Import directly as an idiomatic Go package (`embedded mode`) or run as a standalone worker/dashboard daemon (`standalone mode`).
- **🛡️ First-Class Rate Limiting Synergy (`distlimit`):** Native integration adapter with `github.com/balramadan/distlimit`. Rate-limited jobs are gracefully rescheduled to the Delayed Queue (`RATE_LIMITED` / `DEFERRED`) without consuming retry budgets.
- **⚡ At-Least-Once Atomic Delivery:** Powered by atomic Redis `BLMOVE` operations and crash reclaimers to prevent job loss during worker crashes.
- **🔗 HTTP-Style Middleware Pipeline:** Expressive Chi-like middleware chaining (`Use(...)`) supporting structured logging (`slog`), panic recovery, and rate limiting.
- **📊 Real-time SSE Telemetry:** Built-in Server-Sent Events streamer broadcasting live queue counters (`pending`, `processing`, `delayed`, `dlq`) to a Vue.js web dashboard.

---

## 📊 Competitive Matrix

| Feature / Capability          | **Asynq**       | **River**      | **Machinery**      | **GopherQueue**                                      |
| ----------------------------- | --------------- | -------------- | ------------------ | ---------------------------------------------------- |
| **Primary Broker**            | Redis           | PostgreSQL     | Redis / AMQP / SQS | **Redis** _(Pluggable Storage)_                      |
| **Deployment Mode**           | Embedded Only   | Embedded Only  | Embedded Only      | **Dual-Mode** _(Embedded & Standalone Microservice)_ |
| **Middleware Pipeline**       | Basic           | Basic          | None               | **Full HTTP-Style Pipeline (`Use(...)`)**            |
| **Distributed Rate Limiting** | Basic Key/Limit | Basic          | None               | **Native `distlimit` Integration Adapter**           |
| **Real-time Telemetry**       | Separate Web UI | Postgres Views | None               | **Embedded SSE Streamer + Built-in Vue Dashboard**   |

---

## 📦 Installation

```bash
go get github.com/balramadan/gopherqueue
```

---

## ⚡ Quickstart

### 1. Embedded Producer Client & Worker Server

```go
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
	redisStore := gopherqredis.NewStore(rdb)

	// 1. Initialize Producer Client
	client, _ := gopherqueue.NewClient(
		gopherqueue.WithStorage(redisStore),
		gopherqueue.WithQueue("default"),
	)

	// 2. Initialize Worker Server
	server := gopherqueue.NewServer(
		gopherqueue.WithStorage(redisStore),
		gopherqueue.WithQueue("default"),
		gopherqueue.WithConcurrency(10),
	)

	// 3. Register Job Handler
	server.Register("email:send", func(ctx context.Context, job *gopherqueue.Job) error {
		to, _ := job.Payload["to"].(string)
		fmt.Printf("Sending email to %s (Job ID: %s)\n", to, job.ID)
		return nil
	})

	// 4. Enqueue Job
	job, _ := gopherqueue.NewJob("job-101", "default", "email:send", map[string]any{
		"to": "user@example.com",
	})
	_ = client.Enqueue(context.Background(), job)

	// 5. Run Server Engine
	if err := server.Run(context.Background()); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
```

---

### 2. Integration with `distlimit` Rate Limiter

Protect third-party APIs (OpenAI, Payment Gateways, SMS Services) from rate-limit exhaustion:

```go
package main

import (
	"context"
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
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	redisStore := gopherqredis.NewStore(rdb)

	// Initialize distlimit Rate Limiter (Allow 100 requests per minute)
	limiter, _ := distlimit.New(
		distlimitredis.New(rdb),
		distlimit.WithLimit(100),
		distlimit.WithWindow(1*time.Minute),
		distlimit.WithAlgorithm(tokenbucket.New()),
	)

	server := gopherqueue.NewServer(
		gopherqueue.WithStorage(redisStore),
		gopherqueue.WithConcurrency(10),
	)

	// Attach distlimit Execution Guard Middleware
	server.Use(gopherqdist.New(limiter, func(job *gopherqueue.Job) string {
		return "rate_limit:" + job.Type
	}))

	server.Register("api:call", func(ctx context.Context, job *gopherqueue.Job) error {
		// Handler logic
		return nil
	})

	_ = server.Run(context.Background())
}
```

---

## 🛠️ Middleware Ecosystem

`GopherQueue` provides built-in middlewares under `middleware/`:

- **`middleware/distlimit`**: Execution guard adapter for `github.com/balramadan/distlimit`.
- **`middleware/logger`**: Structured logging using Go standard `log/slog`.
- **`middleware/recover`**: Panic recovery guard preventing worker goroutine crashes.

```go
import (
	gopherqlogger "github.com/balramadan/gopherqueue/middleware/logger"
	gopherqrecover "github.com/balramadan/gopherqueue/middleware/recover"
)

server.Use(gopherqlogger.New(slog.Default()))
server.Use(gopherqrecover.New())
```

---

## 🖥️ Zero-Dependency Web Dashboard GUI

`GopherQueue` features a modern, dark-mode Vue 3 Web Dashboard GUI embedded directly into the Go binary using Go `//go:embed` (`web.DistFS()`).

### Option 1: Standalone Single-Binary Execution (Zero Setup)

No Node.js, `npm`, or `pnpm` required! Developers can launch the Web Dashboard GUI with a single command:

```bash
# 1. Run directly using Go CLI from local workspace
go run ./cmd/gopherq-dashboard

# 2. Or run directly from GitHub
go run github.com/balramadan/gopherqueue/cmd/gopherq-dashboard@latest

# 3. Or install the binary globally to your PATH
go install github.com/balramadan/gopherqueue/cmd/gopherq-dashboard@latest
gopherq-dashboard
```

Open **`http://localhost:8080`** in your browser to inspect real-time metrics (_Pending_, _Processing_, _Delayed_, _DLQ_), active SSE connection status, and job payloads.

### Option 2: Embedded Route in your Go Application

Mount the embedded Vue Dashboard static assets directly inside any Go HTTP router (Gin, Chi, Echo, or standard `net/http`):

```go
package main

import (
	"net/http"

	gopherqredis "github.com/balramadan/gopherqueue/storage/redis"
	"github.com/balramadan/gopherqueue/telemetry"
	"github.com/balramadan/gopherqueue/web"
	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	store := gopherqredis.NewStore(rdb)

	sseBroadcaster := telemetry.NewSSEBroadcaster(store, "default")
	http.HandleFunc("/api/v1/stream", sseBroadcaster.StreamMetrics)

	// Serve the Embedded Vue Web Dashboard GUI
	distFS, _ := web.DistFS()
	http.Handle("/dashboard/", http.StripPrefix("/dashboard/", http.FileServer(http.FS(distFS))))

	http.ListenAndServe(":8080", nil)
}
```

---

## 🤝 Contributing

We welcome contributions from the open-source community! Please review our [CONTRIBUTING.md](CONTRIBUTING.md) guide for guidelines on submitting issues, feature requests, and pull requests.

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
