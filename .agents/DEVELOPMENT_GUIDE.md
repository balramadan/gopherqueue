# DEVELOPMENT_GUIDE.md — Transformation Roadmap & Developer Workflows for GopherQueue

This guide provides step-by-step instructions for developing **GopherQueue** open-source Go package, engine, and real-time Web Dashboard GUI.

---

## Transformation Roadmap

```
  Phase 1: Core Package Refactoring & Root Types
  ├── Move domain structs to repository root (`job.go`, `client.go`, `server.go`, `options.go`)
  └── Define Middleware Contract (`middleware.go`)

  Phase 2: Storage Layer Abstraction (`storage/`)
  ├── Define `storage.Engine` interface (`storage/storage.go`)
  └── Implement Redis storage driver (`storage/redis/`) with `Reschedule()` method for rate limiting

  Phase 3: Middleware Ecosystem & `distlimit` Integration Adapter
  ├── Implement logger middleware (`middleware/logger/`) & panic recovery (`middleware/recover/`)
  └── Implement `distlimit` rate limiter adapter (`middleware/distlimit/`)

  Phase 4: Telemetry & Standalone Applications (`cmd/`)
  ├── Build SSE Broadcaster & Telemetry (`telemetry/sse.go`, `telemetry/metrics.go`)
  └── Build standalone server (`cmd/gopherq-server`) and standalone dashboard (`cmd/gopherq-dashboard`)

  Phase 5: Examples, Documentation & Quality Assurance
  ├── Add idiomatic examples under `_examples/`
  └── Comprehensive Unit & Integration Tests (`go test -v -race ./...`)

  Phase 6: Web Dashboard GUI Modernization (`web/`)
  ├── Modern UI/UX Design (Vue 3 + TailwindCSS + Dark Mode Glassmorphism)
  ├── Interactive Job Explorer & Real-Time Throughput Charts
  └── Go Single-Binary Embedding (`embed.FS`) in `cmd/gopherq-dashboard`
```

---

## Guidelines for Web Dashboard GUI Development (`web/`)

1. **Design Aesthetics & UI/UX Principles**:
   - **Modern Aesthetic**: Dark mode by default, glassmorphism card surfaces, clean typography (Inter / Outfit).
   - **Real-Time Visual Indicators**: Active SSE connection status pill (`LIVE`, `RECONNECTING`, `DISCONNECTED`).
   - **Metric Cards & Charts**: Real-time throughput cards for Pending, Processing, Delayed, and Dead Letter Queue (DLQ).
   - **Interactive Job Explorer**: Filterable table to inspect job payloads, status badges, retry counts, error stack traces, and manual requeue/delete controls.

2. **Single-Binary Embedding (`embed.FS`)**:
   - Compile Vue.js assets to `web/dist/`.
   - Embed `web/dist` in `cmd/gopherq-dashboard/main.go` using `//go:embed dist/*` so users can run the dashboard as a single standalone Go binary without needing Node.js or static file servers.

3. **Backend Telemetry Integration**:
   - Consume SSE stream from `/api/v1/stream`.
   - Maintain client-side reconnection strategy with exponential backoff.

---

## Guidelines for AI Agents

1. **Root Package Exposure**:
   - Primary user-facing components (`Job`, `Server`, `Client`, `MiddlewareFunc`) must reside in `package gopherqueue` at the root level for idiomatic `go get` imports.

2. **Error Handling & Types**:
   - Define custom error types like `ErrRateLimited` in `job.go` so middleware adapters can communicate rescheduling requests back to the engine cleanly.

3. **Verification Standards**:
   - Ensure all public functions have descriptive GoDoc comments in English.
   - Run `go test -v -race ./...` and `go vet ./...` before finishing any implementation task.
