package gopherqueue

import (
	"context"
	"errors"
	"log"
	"sync"
)

// Server represents the main GopherQueue worker engine orchestrator.
type Server struct {
	cfg      *Config
	registry *HandlerRegistry
	pool     *WorkerPool
	chain    Chain
	mu       sync.Mutex
	running  bool
}

// NewServer initializes a new Server instance with functional options.
func NewServer(opts ...Option) *Server {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	return &Server{
		cfg:      cfg,
		registry: NewHandlerRegistry(),
	}
}

// Register registers a job handler function for a specific job type.
func (s *Server) Register(jobType string, fn func(ctx context.Context, job *Job) error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.registry.Register(jobType, func(ctx context.Context, job *Job) error {
		pipeline := s.chain.Then(HandlerFunc(fn))
		return pipeline.ProcessJob(ctx, job)
	})
}

// Use appends one or more middlewares to the server's global middleware pipeline.
func (s *Server) Use(middlewares ...MiddlewareFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chain = s.chain.Append(middlewares...)
}

// Run starts the worker pool and blocks until the context is cancelled or Shutdown is called.
func (s *Server) Run(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return errors.New("gopherqueue: server is already running")
	}
	if s.cfg.Storage == nil {
		s.mu.Unlock()
		return errors.New("gopherqueue: storage engine is required")
	}
	s.running = true
	s.mu.Unlock()

	log.Printf("[GopherQueue Server] Starting engine on queue '%s' with %d workers...", s.cfg.QueueName, s.cfg.Concurrency)

	s.pool = NewWorkerPool(s.cfg.QueueName, s.cfg.Concurrency, s.cfg.PollInterval, s.cfg.Storage, s.registry)
	s.pool.Start(ctx)

	<-ctx.Done()
	return s.Shutdown(context.Background())
}

// Shutdown gracefully stops all active workers.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	log.Println("[GopherQueue Server] Shutting down worker engine...")
	if s.pool != nil {
		s.pool.Stop()
	}
	s.running = false
	log.Println("[GopherQueue Server] Server stopped cleanly.")
	return nil
}
