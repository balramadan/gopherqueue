# DESIGN.md — Architecture & Technical Specifications for GopherQueue

This document describes the architectural specifications, state machine transitions, storage key design, middleware pipeline, `distlimit` (`github.com/balramadan/distlimit`) integration mechanics, and **Web Dashboard GUI Design System** for **GopherQueue**.

---

## 1. Job Lifecycle & State Machine

```
                  +---------------+
                  |    ENQUEUED   |
                  +-------+-------+
                          |
             +------------+------------+
             |                         | (delayed)
             v                         v
      +--------------+          +--------------+
      |   PENDING    |          |   DELAYED    |
      +------+-------+          +------+-------+
             |                         | (timer elapsed)
             +------------+------------+
                          |
                          v (Worker fetches job via BLMOVE)
                  +---------------+
                  |  PROCESSING   |
                  +-------+-------+
                          |
         +----------------+----------------+----------------+
         | (Success)                       | (Failure)      | (Rate Limited)
         v                                 v                v
  +--------------+                  +--------------+ +--------------+
  |  COMPLETED   |                  |    FAILED    | | RATE_LIMITED |
  +--------------+                  +------+-------+ +------+-------+
                                           |                |
                       +-------------------+                | (Reschedule without incrementing retry count)
                       | (Attempts < Max)                   v
                       v                           +----------------+
                +--------------+                   |    DELAYED     |
                |   RETRYING   |                   +----------------+
                +--------------+
                       | (Attempts >= Max)
                       v
                +--------------+
                |  DEAD_LETTER |
                +--------------+
```

### State Machine Transition Rules

| State Transition                                                | Trigger / Logic                                                                                                                                            |
| --------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ENQUEUED` $\rightarrow$ `PENDING`                              | Immediate job enqueue (`ExecuteAt <= Now`).                                                                                                                |
| `ENQUEUED` $\rightarrow$ `DELAYED`                              | Scheduled job enqueue (`ExecuteAt > Now`).                                                                                                                 |
| `DELAYED` $\rightarrow$ `PENDING`                               | `DelayedScheduler` polls ZSet (`Score <= Now`) and moves job ID to pending list.                                                                           |
| `PROCESSING` $\rightarrow$ `COMPLETED`                          | Handler completes execution without error; job is ACKed.                                                                                                   |
| `PROCESSING` $\rightarrow$ `RATE_LIMITED`                       | `distlimit` middleware returns `ErrRateLimited`. Job is moved back to `DELAYED` queue with `ExecuteAt = Now + RetryAfter`. **No retry counter increment**. |
| `PROCESSING` $\rightarrow$ `FAILED` $\rightarrow$ `RETRYING`    | Handler returns execution error (`CurrentRetry < MaxRetries`). Job is moved to `DELAYED` queue with exponential backoff delay.                             |
| `PROCESSING` $\rightarrow$ `FAILED` $\rightarrow$ `DEAD_LETTER` | Handler returns execution error (`CurrentRetry >= MaxRetries`). Job is moved to `DLQ`.                                                                     |

---

## 2. Refactored Directory Structure & Package Responsibilities

```text
gopherqueue/
├── client.go                      # Enqueue Client API (Producer)
├── server.go                      # Worker Server & Engine Orchestrator
├── job.go                         # Domain Model (Job, JobStatus, ErrRateLimited)
├── middleware.go                  # Middleware Pipeline Engine (Chi-like pattern)
├── options.go                     # Functional Options Configuration
├── worker.go                      # Worker Pool & Reschedule Engine
├── doc.go                         # Package-level GoDoc Documentation
│
├── middleware/                    # Built-in Middlewares
│   ├── logger/                    # Structured Logging (slog)
│   ├── recover/                   # Panic Recovery Guard
│   └── distlimit/                 # Rate Limiter Adapter (distlimit integration)
│
├── storage/                       # Storage Abstraction Layer
│   ├── storage.go                 # Engine Interface Specification
│   └── redis/                     # Redis Driver Implementation
│
├── telemetry/                     # Metrics & Real-time Telemetry
│   ├── sse.go                     # Server-Sent Events Broadcaster
│   └── metrics.go                 # Metrics Snapshot Collector
│
├── web/                           # Vue.js 3 Web Dashboard Frontend
│   └── src/
│       ├── components/            # Reusable UI Components
│       ├── views/                 # View Pages (Metrics, Explorer, Settings)
│       └── assets/                # Design System & Tailwind CSS Token Setup
│
├── cmd/                           # Executable Entry Points
│   ├── gopherq-server/            # Standalone Ingestion & Worker Daemon
│   └── gopherq-dashboard/         # Standalone Web Dashboard Server
│
├── _examples/                     # Usage Examples & Demonstration Code
├── go.mod
├── Makefile
└── README.md
```

---

## 3. Storage Abstraction Layer & Redis Key Design

### Storage Interface Contract (`storage/storage.go`)

```go
type Storage interface {
	Enqueue(ctx context.Context, job *Job) error
	Fetch(ctx context.Context, queueName string, timeout time.Duration) (*Job, error)
	Ack(ctx context.Context, job *Job) error
	Nack(ctx context.Context, job *Job, execErr error) error
	Reschedule(ctx context.Context, job *Job, retryAfter time.Duration) error
	GetPendingCount(ctx context.Context, queueName string) (int64, error)
	GetProcessingCount(ctx context.Context, queueName string) (int64, error)
	GetDelayedCount(ctx context.Context, queueName string) (int64, error)
	GetDLQCount(ctx context.Context, queueName string) (int64, error)
}
```

### Redis Key Schema (`storage/redis/`)

| Redis Key Structure                     | Type              | Description                                            |
| --------------------------------------- | ----------------- | ------------------------------------------------------ |
| `gopherq:queue:{queue_name}:pending`    | **List**          | FIFO queue of pending job IDs.                         |
| `gopherq:queue:{queue_name}:processing` | **List**          | Unacknowledged job IDs currently processing.           |
| `gopherq:queue:{queue_name}:delayed`    | **ZSet**          | Delayed & deferred job IDs with Unix timestamp score.  |
| `gopherq:queue:{queue_name}:dlq`        | **List**          | Permanently failed job IDs.                            |
| `gopherq:job:{job_id}`                  | **String (JSON)** | Full JSON representation and status metadata of a job. |

---

## 4. `distlimit` Middleware Integration Mechanics

1. `gopherqueue/middleware/distlimit` wraps `distlimit.RateLimiter` (`github.com/balramadan/distlimit`).
2. Before calling `next.ProcessJob(ctx, job)`, the middleware evaluates `limiter.Allow(ctx, key)`.
3. If allowed, execution continues to the handler.
4. If rate-limited (`allowed == false`), the middleware returns `&ErrRateLimited{RetryAfter: d, Key: key}`.
5. The `Server` worker engine catches `ErrRateLimited` and delegates to `storage.Reschedule(ctx, job, retryAfter)`, moving the job to `queue:delayed` without incrementing `CurrentRetry`.

---

## 5. Web Dashboard GUI Design System & UI/UX Standards (`web/`)

### Color Palette (Curated Dark Mode)

| Category / Element | Token Name | Hex Code | Description / Usage |
| --- | --- | --- | --- |
| **App Background** | `bg-main` | `#0F172A` (Slate 900) | Main dashboard viewport background. |
| **Card Surface** | `bg-surface` | `#1E293B` (Slate 800) | Glassmorphism card surface with `border: 1px solid rgba(255,255,255,0.08)`. |
| **Primary Accent** | `accent-brand` | `#6366F1` (Indigo 500) | Buttons, active navigation, interactive highlights. |
| **Pending Status** | `status-pending` | `#F59E0B` (Amber 500) | Pending jobs count & badges. |
| **Processing Status**| `status-processing`| `#06B6D4` (Cyan 500) | Active processing jobs count & spinners. |
| **Delayed Status** | `status-delayed` | `#8B5CF6` (Purple 500) | Scheduled / Rate-limited deferred jobs count. |
| **Completed Status**| `status-completed` | `#10B981` (Emerald 500)| Successfully ACKed jobs count & badges. |
| **DLQ Status** | `status-dlq` | `#EF4444` (Rose 500) | Dead Letter Queue (DLQ) alert count & error badges. |

### Reusable UI Components Architecture (`web/src/components/`)

AI agents contributing to the Vue dashboard **MUST** build modular, reusable components:

1. **`MetricCard.vue`**:
   - Reusable card component displaying a metric title, numeric value, icon, and status glow accent.
2. **`StatusBadge.vue`**:
   - Reusable badge displaying `JobStatus` (`PENDING`, `PROCESSING`, `RATE_LIMITED`, `COMPLETED`, `DEAD_LETTER`) with corresponding color token.
3. **`ConnectionPill.vue`**:
   - Live status indicator pill displaying real-time SSE stream status: `LIVE` (Emerald), `RECONNECTING` (Amber), `DISCONNECTED` (Rose).
4. **`JobExplorerTable.vue`**:
   - Filterable data table with search bar, pagination, queue filter dropdown, and action buttons (*Requeue*, *Delete*).
5. **`JobDetailModal.vue`**:
   - Drawer/Modal displaying raw JSON payload syntax highlighting and execution error stack traces.

### Design Consistency & Responsiveness Rules

- **Typography**: Inter or Outfit sans-serif fonts.
- **Grid Layout**: Responsive grid (`grid-cols-1 md:grid-cols-2 lg:grid-cols-4`).
- **Micro-Animations**: Smooth transitions on hover (`transition-all duration-200 ease-in-out`), subtle pulse animation on SSE live data updates.
