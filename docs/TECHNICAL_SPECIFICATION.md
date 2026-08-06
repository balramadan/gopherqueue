# Technical Specification Document v0.1

## **GopherQueue** — Open-Source Distributed Task Processing Library & Engine for Go

---

## 1. Executive Summary & Unique Value Proposition (UVP)

**GopherQueue v0.1** ditransformasi dari sekadar _standalone microservice_ menjadi **Go Open-Source Library/Package** yang dapat di-_import_ langsung ke dalam aplikasi Go (`embedded mode`) sekaligus dapat dijalankan sebagai _standalone engine_ (`service mode`).

### Competitive Matrix

| Fitur / Kapabilitas           | **Asynq**       | **River**      | **Machinery**      | **GopherQueue**                                        |
| ----------------------------- | --------------- | -------------- | ------------------ | ------------------------------------------------------ |
| **Primary Broker**            | Redis           | PostgreSQL     | Redis / AMQP / SQS | **Redis** _(Pluggable Storage)_                        |
| **Deployment Mode**           | Embedded Only   | Embedded Only  | Embedded Only      | **Dual-Mode** _(Embedded & Standalone Microservice)_   |
| **Middleware Architecture**   | Ada (Basic)     | Ada            | Tidak Ada          | **Full Middleware Pipeline** _(HTTP-style / Chi-like)_ |
| **Distributed Rate Limiting** | Basic Key/Limit | Basic          | Tidak Ada          | **Native `distlimit` Integration Adapter**             |
| **Real-time Telemetry**       | Separate Web UI | Postgres Views | None               | **Embedded SSE Streamer + Built-in Vue Dashboard**     |

### Key Value Propositions (UVP)

1. **Dual-Mode Flexibility**: Pengembang dapat memilih meng-import `GopherQueue` langsung di dalam monolit Go mereka _atau_ memisahkan pengeksekusian _job_ ke kluster _worker dedicated_ tanpa mengubah struktur kode _payload_.
2. **First-Class Rate Limiting Synergy (`distlimit`)**: Mengintegrasikan `distlimit` (`github.com/balramadan/distlimit`) sebagai _execution guard middleware_ berkinerja tinggi (sub-100ns, zero-alloc `0 B/op`, hybrid circuit-breaker). Pengeksekusian job yang terbentur kuota rate limit tidak dianggap sebagai _retry failure_, melainkan secara anggun di-reschedule (_graceful deferred reschedule_) ke _Delayed Queue_.
3. **HTTP-Style Middleware Pipeline**: Ekosistem _middleware_ yang familier bagi pengembang Go (mendukung logging, recovery, distributed tracing, dan rate limiting).
4. **Zero-Dependency SSE Broadcaster**: Monitoring status antrean secara _real-time_ tanpa perlu mengkonfigurasi server WebSocket terpisah.

---

## 2. Refactored Package Architecture

Agar dapat di-import oleh proyek Go lain (`go get github.com/balramadan/gopherqueue`), folder `internal/` dirombak total menjadi struktur publik yang idiomatik.

```text
gopherqueue/
├── client.go                      # High-level Producer Client (Enqueue API)
├── server.go                      # Worker Server & Engine Orchestrator
├── job.go                         # Core Domain Model (Job, Status, Payload, ErrRateLimited)
├── middleware.go                  # Middleware Contract & Pipeline Engine
├── options.go                     # Functional Options Pattern Config
│
├── middleware/                    # Built-in Middlewares
│   ├── logger/                    # Structured Logging Middleware (slog)
│   ├── recover/                   # Panic Recovery Guard
│   └── distlimit/                 # Rate Limiter Adapter (Integration with distlimit)
│
├── storage/                       # Abstraksi Storage Layer (Pluggable)
│   ├── storage.go                 # Engine Interface Specification
│   └── redis/                     # Redis Driver Implementation (BLMOVE, ZSet)
│
├── telemetry/                     # Monitoring & Telemetry
│   ├── sse.go                     # Server-Sent Events Broadcaster
│   └── metrics.go                 # Prometheus Metrics Exporter
│
├── cmd/                           # Executables (Standalone Mode)
│   ├── gopherq-server/            # Standalone Worker & Ingestion Daemon
│   └── gopherq-dashboard/         # Standalone Web Dashboard Server
│
├── _examples/                     # Code Examples & Best Practices
├── go.mod
└── README.md
```

---

## 3. Integrasi & Sinergi Ekosistem: `GopherQueue` + `distlimit`

### Context: Mengenai Library `distlimit`

`distlimit` (`github.com/balramadan/distlimit`) adalah pustaka _distributed rate limiting_ untuk Go berkinerja tinggi (evaluasi sub-100 nanolatihan, _zero memory allocation_ `0 B/op`). `distlimit` mendukung 5 algoritma utama:

- **Token Bucket**: Penanganan burst traffic umum.
- **Leaky Bucket**: _Traffic shaping_ untuk integrasi pihak ketiga (Payment Gateway, SMS Broker).
- **Fixed Window**: _Counter interval reset_ berbobot ringan.
- **Sliding Window Log**: Presisi 100% menggunakan Sorted Set (ZSET).
- **Sliding Window Counter**: Presisi _moving average_ berkinerja tinggi $O(1)$.

Selain itu, `distlimit` memiliki **Hybrid Storage Driver (Redis + Memory Fallback)** dengan _Half-Open Circuit Breaker_, sehingga saat jaringan Redis terganggu, sistem tidak meledak melainkan secara otomatis _failover_ ke in-memory storage.

### Problem Statement

Banyak _background job_ (seperti pengiriman email, ekstraksi LLM OpenAI, pemanggilan Payment Gateway, SMS API) memiliki batas kuota rate limit yang ketat dari penyedia layanan. Jika _worker_ GopherQueue mengambil job namun eksekusinya gagal akibat terbentur _rate limit_ API pihak ketiga:

1. Job dianggap gagal (`FAILED`) dan mengkonsumsi/menghabiskan jatah `MaxRetries`.
2. Terjadi lonjakan pemanggilan berulang (_Thundering Herd Problem_).

### Solution Mechanics: `distlimit` Execution Guard Adapter

`distlimit` diintegrasikan melalui paket adapter `github.com/balramadan/gopherqueue/middleware/distlimit` sebagai **Execution Guard Middleware** di dalam _Worker Pipeline_.

```
                                [ Worker Pool ]
                                       |
                            (Fetches Job from Queue)
                                       |
                                       v
                    +------------------------------------+
                    |   gopherq/middleware/distlimit     |
                    +------------------------------------+
                                       |
                            Evaluasi distlimit.Allow()
                                       |
                     +-----------------+-----------------+
                     |                                   |
           [ Rate Limit Exceeded ]                 [ Rate Limit OK ]
                     |                                   |
                     v                                   v
      +-----------------------------+          +-------------------+
      |   DEFERRED RESCHEDULE       |          |   EXECUTE JOB     |
      | - NO increment to Retries   |          |   PAYLOAD         |
      | - Move to Delayed Queue     |          +-------------------+
      |   with Retry-After Backoff  |
      +-----------------------------+
```

### State Machine Adjustment: `RATE_LIMITED` / `DEFERRED`

Diperkenalkan status pendukung baru: **`RATE_LIMITED`**.

- Jika `distlimit.Allow(ctx, key)` mengembalikan nilai `allowed == false`:

1. Middleware mengembalikan error khusus `&gopherqueue.ErrRateLimited{RetryAfter: retryAfter, Key: limitKey}`.
2. Engine **TIDAK** menambah `CurrentRetry` (_retry counter_ utuh).
3. Engine membaca `RetryAfter` dari error `ErrRateLimited` (misal 5 detik).
4. Engine memanggil `storage.Defer(ctx, job, retryAfter)` untuk memindahkan _job_ kembali ke **Delayed Queue** dengan score $\text{execute\_at} = \text{now} + \text{RetryAfter}$.

---

## 4. Code Design & Public API (Developer Experience)

### 4.1 Implementation Adapter `middleware/distlimit`

```go
package distlimit

import (
	"context"
	"fmt"
	"time"

	"github.com/balramadan/distlimit" // Integration with distlimit package
	"github.com/balramadan/gopherqueue"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, time.Duration, error)
}

// NewMiddleware membuat gopherqueue.MiddlewareFunc menggunakan engine distlimit.
func NewMiddleware(limiter RateLimiter, keyExtractor func(job *gopherqueue.Job) string) gopherqueue.MiddlewareFunc {
	return func(next gopherqueue.Handler) gopherqueue.Handler {
		return gopherqueue.HandlerFunc(func(ctx context.Context, job *gopherqueue.Job) error {
			limitKey := keyExtractor(job)

			allowed, retryAfter, err := limiter.Allow(ctx, limitKey)
			if err != nil {
				// Fallback strategy jika rate limiter error: Log & Lanjutkan atau Delay
				return fmt.Errorf("distlimit middleware error: %w", err)
			}

			if !allowed {
				// Signal khusus ke GopherQueue Engine bahwa job ini ter-Rate Limit
				return &gopherqueue.ErrRateLimited{
					RetryAfter: retryAfter,
					Key:        limitKey,
				}
			}

			return next.ProcessJob(ctx, job)
		})
	}
}
```

---

### 4.2 Pemakaian oleh Developer (Idiomatic Library Usage)

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/balramadan/distlimit"
	"github.com/balramadan/distlimit/algorithm/tokenbucket"
	distlimitredis "github.com/balramadan/distlimit/driver/redis"
	"github.com/balramadan/gopherqueue"
	gopherdist "github.com/balramadan/gopherqueue/middleware/distlimit"
)

func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	// 1. Inisialisasi Distlimit Engine (Rate Limiter)
	redisDriver := distlimitredis.New(rdb)
	limiter, err := distlimit.New(
		redisDriver,
		distlimit.WithLimit(100),
		distlimit.WithWindow(1*time.Minute),
		distlimit.WithAlgorithm(tokenbucket.New()),
	)
	if err != nil {
		log.Fatalf("Failed to init distlimit: %v", err)
	}

	// 2. Inisialisasi GopherQueue Server
	server := gopherqueue.NewServer(
		gopherqueue.WithRedis(rdb),
		gopherqueue.WithConcurrency(10),
		gopherqueue.WithQueue("default"),
	)

	// 3. Pasang Middleware Rate Limiting (distlimit)
	server.Use(gopherdist.NewMiddleware(limiter, func(job *gopherqueue.Job) string {
		// Rate limit berdasarkan Tipe Job / Tenant ID
		return "rate_limit:" + job.Type
	}))

	// 4. Register Handler
	server.Register("email:send", func(ctx context.Context, job *gopherqueue.Job) error {
		log.Printf("Executing email job: %s", job.ID)
		return nil
	})

	// 5. Jalankan Engine
	if err := server.Run(context.Background()); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
```

---

## 5. Areas for Improvement & Strategic Roadmap

Meskipun keunggulan _Rate Limiting_ dan _Dual-Mode_ sudah memberikan posisi tawar tinggi dibanding kompetitor, ada beberapa area pengembangan jangka panjang:

### 1. Unique Jobs (Deduplication)

- **Tantangan**: Mencegah _job_ identik dimasukkan ke antrean lebih dari sekali selama durasi tertentu.
- **Solusi**: Menyediakan opsi `WithUnique(ttl)` saat _Enqueue_ dengan menyimpan hash payload di Redis Key bertipe String dengan TTL.

### 2. Workflow / Task Chaining (DAG Execution)

- **Tantangan**: Menjalankan _job_ berurutan (misal: `Job B` baru berjalan jika `Job A` sukses, atau `Group Job` yang berjalan secara paralel lalu di-aggregate).
- **Solusi**: Merancang pustaka tingkat lanjut `gopherqueue/workflow` untuk mendukung arsitektur Directed Acyclic Graph (DAG).

### 3. Multi-Broker Storage Abstraction

- **Tantangan**: Beberapa perusahaan menghindari Redis karena masalah _memory footprint_ dan memilih PostgreSQL atau NATS JetStream.
- **Solusi**: Memperjelas interface `storage.Engine` agar pengembang open-source dapat membuat driver `storage/postgres` atau `storage/nats`.

### 4. Long-Term Auditing & History Storage

- **Tantangan**: Redis tidak ideal untuk menyimpan riwayat _job_ jutaan transaksi dalam hitungan bulan.
- **Solusi**: Membuat fitur _Archive Hook_ yang secara otomatis mentransfer data _job_ bernilai `COMPLETED` / `DEAD_LETTER` dari Redis ke SQL Database (PostgreSQL/MySQL) atau Time-Series Storage.
