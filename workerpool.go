package workerpool

import (
	"errors"
	"sync"
	"time"
)

type Pool struct {
	MaxCapacity int
	Capacity    int
	workers     chan *Worker
	workerCache sync.Pool
	closing     bool
	closechan   chan struct{}
}

var (
	ErrPoolClosed     = errors.New("this pool has been closed")
	ErrWorkerShortage = errors.New(" worker amount cannot be lower than zero")
	ErrWorkerExecss   = errors.New(" worker amount cannot be greater than maxcapacity")
)

func NewPool(maxWorkers int, duration time.Duration) *Pool {
	// There must be at least one worker.
	if maxWorkers < 1 {
		maxWorkers = 1
	}

	pool := &Pool{
		MaxCapacity: maxWorkers,
		Capacity:    0,
		workers:     make(chan *Worker, maxWorkers),
		closechan:   make(chan struct{}),
	}
	pool.workerCache.New = func() interface{} {
		return &Worker{
			pool: pool,
			task: make(chan func()),
		}
	}
	pool.addWorker(maxWorkers)

	pm := &PoolManager{
		CheckDuration:  time.Duration(duration),
		minWorkerRatio: 0.2,
		decreaseRatio:  0.3,
		increaseRatio:  0.8,
	}

	go pm.Manage(pool)
	return pool
}

func (pool *Pool) releaseWorker(num int) error {
	if num > pool.Capacity {
		return ErrWorkerShortage
	}
	pool.Capacity -= num
	for i := 0; i < num; i++ {
		w := <-pool.workers
		w.task <- nil
	}
	return nil
}

func (pool *Pool) addWorker(num int) error {
	if pool.MaxCapacity < pool.Capacity+num {
		return ErrWorkerExecss
	}
	pool.Capacity += num
	for i := 0; i < num; i++ {
		w := pool.workerCache.Get().(*Worker)
		w.run()
		pool.workers <- w
	}
	return nil
}

func (pool *Pool) Submit(task func()) error {
	if pool.closing {
		return ErrPoolClosed
	}
	w := <-pool.workers
	w.task <- task
	return nil
}

func (pool *Pool) revertWorker(w *Worker) {
	pool.workers <- w
}

func (pool *Pool) close() error {
	pool.releaseWorker(pool.Capacity)
	close(pool.workers)
	close(pool.closechan)
	return nil
}

func (pool *Pool) SetClose() {
	pool.closing = true
}

func (pool *Pool) WaitClose() {
	pool.closing = true
	<-pool.closechan
}
