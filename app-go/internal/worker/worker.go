package worker

import (
	"context"
	"runtime"
	"time"
)

// WorkerPool defines the contract for acquiring and releasing bounded concurrency slots.
type WorkerPool interface {
	Acquire(ctx context.Context) error
	Release()
	Capacity() int
	InUse() int
}

type channelPool struct {
	sem chan struct{}
	cap int
}

// NewBoundedPool initializes a semaphore-based bounded worker pool.
func NewBoundedPool(capacity int) WorkerPool {
	if capacity < 1 {
		capacity = 1
	}
	return &channelPool{
		sem: make(chan struct{}, capacity),
		cap: capacity,
	}
}

func (p *channelPool) Acquire(ctx context.Context) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}
	select {
	case p.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *channelPool) Release() {
	select {
	case <-p.sem:
	default:
	}
}

func (p *channelPool) Capacity() int {
	return p.cap
}

func (p *channelPool) InUse() int {
	return len(p.sem)
}

// PoolManager holds dedicated worker pools for CPU and Allocation tasks.
// Note: Database connection pooling is managed natively by database/sql.
type PoolManager struct {
	CpuPool   WorkerPool
	AllocPool WorkerPool
}

// NewPoolManager initializes dedicated CPU and Allocation worker pools bounded by CPU cores.
func NewPoolManager() *PoolManager {
	numCPU := runtime.GOMAXPROCS(0)
	if numCPU < 1 {
		numCPU = 1
	}

	return &PoolManager{
		CpuPool:   NewBoundedPool(numCPU), // Bounded by CPU Cores (W1 CPU Bound)
		AllocPool: NewBoundedPool(numCPU), // Bounded by Allocation Cores (W4 Allocation)
	}
}
