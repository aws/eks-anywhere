package workerpool

// Workload defines a workload to be run by the worker pool.
type Workload func()

// PoolOption describes a configurable option for the worker pool.
type PoolOption func(*Config)

// Config represents the configuration for the worker pool.
type Config struct {
	// Number of workers in the pool.
	Count int

	// Buffer size of the work queue.
	Buffer int
}

// Pool represents a pool of workers that can execute arbitrary workloads concurrently.
type Pool struct {
	feed chan Workload
}

// Spawn constructs a Pool instance with 10 workers and a buffer size of 0. This means calls to
// Run() will block until a worker is available.
func Spawn(opts ...PoolOption) Pool {
	cfg := Config{
		Count: 10,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	feed := make(chan Workload, cfg.Buffer)
	for i := 0; i < cfg.Count; i++ {
		go func() {
			for w := range feed {
				w()
			}
		}()
	}

	return Pool{feed: feed}
}

// Run runs a workload in the worker pool. It should only be called after Start(). If Run() is
// called after the pool is stopped, it will panic.
func (p Pool) Run(w Workload) {
	p.feed <- w
}

// Stop stop the worker pool. It should only be called after Start(). Once stopped, the worker
// pool should not have Workload's submitted to it.
func (p Pool) Stop() {
	close(p.feed)
}

// Workers configures the number of workers in the pool.
func Workers(count int) PoolOption {
	return func(cfg *Config) {
		cfg.Count = count
	}
}

// Buffer configures the buffer size of the work queue. Sizes > 0 will allow calls to Run()
// without blocking until the buffer is full.
func Buffer(size int) PoolOption {
	return func(cfg *Config) {
		cfg.Buffer = size
	}
}
