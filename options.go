package gopherqueue

import (
	"time"
)

// Config holds configuration parameters for Server and Client instances.
type Config struct {
	QueueName      string
	Concurrency    int
	PollInterval   time.Duration
	DefaultTimeout time.Duration
	Storage        Storage
}

func defaultConfig() *Config {
	return &Config{
		QueueName:      "default",
		Concurrency:    10,
		PollInterval:   2 * time.Second,
		DefaultTimeout: 30 * time.Second,
	}
}

// Option represents a functional option for configuring Server or Client.
type Option func(*Config)

// WithQueue sets the queue name to listen to or push jobs into.
func WithQueue(name string) Option {
	return func(c *Config) {
		if name != "" {
			c.QueueName = name
		}
	}
}

// WithConcurrency sets the number of concurrent worker goroutines.
func WithConcurrency(n int) Option {
	return func(c *Config) {
		if n > 0 {
			c.Concurrency = n
		}
	}
}

// WithPollInterval sets the interval between worker queue polling calls.
func WithPollInterval(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.PollInterval = d
		}
	}
}

// WithStorage sets the storage engine implementation (e.g. redis.Store).
func WithStorage(s Storage) Option {
	return func(c *Config) {
		if s != nil {
			c.Storage = s
		}
	}
}
