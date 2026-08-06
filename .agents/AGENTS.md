# AGENTS.md — AI Agent Execution Rules & Coding Guidelines for GopherQueue

Welcome, AI Agent! This document defines the core principles, operational constraints, and architectural standards for contributing to **GopherQueue**.

---

## 1. System Overview

**GopherQueue** is an open-source distributed task processing library and background job engine written in Go. It operates in **Dual-Mode**:

- **Embedded Mode**: Directly imported as a Go package (`go get github.com/balramadan/gopherqueue`) inside monolith or microservice Go applications.
- **Standalone Mode**: Executed as a standalone server (`cmd/gopherq-server`) and web dashboard (`cmd/gopherq-dashboard`).

### Key Value Propositions

- **At-Least-Once Delivery**: Atomic `BLMOVE` operations prevent job loss during worker crashes.
- **First-Class Rate Limiting Synergy (`distlimit`)**: Integrated `distlimit` (`github.com/balramadan/distlimit`) middleware gracefully defers rate-limited jobs to the delayed queue without consuming retry budgets.
- **HTTP-Style Middleware Pipeline**: Familiar middleware chaining (`Use(middleware...)`) supporting logging, panic recovery, distributed tracing, and rate-limiting.
- **Pluggable Storage Abstraction**: Decoupled `storage.Engine` interface with Redis as the default primary driver.
- **Zero-Dependency SSE Broadcaster & Dashboard GUI**: Embedded Server-Sent Events (SSE) telemetry stream paired with a Vue.js web dashboard.

---

## 2. Core Rules for AI Agents

### Rule 1: Zero Job Loss & State Invariants

- Use atomic Redis operations (`BLMOVE`, Lua scripts).
- Jobs must remain in the `processing` state until confirmed `COMPLETED` (ACK), gracefully rescheduled (`DEFERRED`/`RATE_LIMITED`), or moved to `DLQ`.
- Never silently drop or ignore Redis errors.

### Rule 2: Rate Limiting & Graceful Rescheduling

- When `distlimit` middleware returns `ErrRateLimited`, the engine **MUST NOT** increment `CurrentRetry`.
- Reschedule the job to `queue:delayed` with score `execute_at = now + retry_after`.

### Rule 3: Public API & Package Layout Conventions

- Maintain idiomatic Go package layout: core user-facing types reside at the repository root (`client.go`, `server.go`, `job.go`, `middleware.go`, `options.go`).
- Middlewares belong under `middleware/` (e.g. `middleware/distlimit`, `middleware/logger`).
- Storage drivers belong under `storage/` (e.g. `storage/redis`).
- Telemetry & Dashboard belong under `telemetry/`, `web/`, and `cmd/` (`cmd/gopherq-server`, `cmd/gopherq-dashboard`).

### Rule 4: Concurrency & Graceful Shutdown

- Goroutines must respect `context.Context` cancellation and timeouts.
- Implement graceful shutdown trapping `SIGTERM` and `SIGINT`, draining active jobs within configurable grace periods.

### Rule 5: Web Dashboard GUI Standards (`web/`)

- Use Vue 3 + Vite + TailwindCSS with modern dark-mode aesthetic.
- Connect to SSE stream at `/api/v1/stream` with real-time status indicators and error handlers.
- Support Go single-binary embedding via `embed.FS` in `cmd/gopherq-dashboard`.

---

## 3. Verification & Quality Gates

Before finalizing any task, AI agents MUST verify:

1. `go test -v -race ./...` passes without data races or errors.
2. Build commands (`go build ./...`) compile cleanly.
3. Public signatures use clear docstrings and functional options patterns.
