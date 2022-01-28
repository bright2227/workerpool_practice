package workerpool

import (
	"testing"
	"time"
)

func TestExample(t *testing.T) {
	pool := NewPool(2, 1*time.Second)
	requests := []string{"alpha", "beta", "gamma", "delta", "epsilon"}

	rspChan := make(chan string, len(requests))
	for _, r := range requests {
		r := r
		pool.Submit(func() {
			rspChan <- r
		})
	}
	pool.WaitClose()
	close(rspChan)

	for r := range rspChan {
		t.Log(r)
	}
}

func TestPanic(t *testing.T) {
	pool := NewPool(2, 1*time.Second)
	requests := []string{"alpha", "beta", "gamma", "delta", "epsilon"}

	rspChan := make(chan string, len(requests))
	for _, r := range requests {
		r := r
		pool.Submit(func() {
			rspChan <- r
			panic(" test panic, worker don't lose ")
		})
	}
	pool.WaitClose()
	close(rspChan)

	for r := range rspChan {
		t.Log(r)
	}
}

func TestGoroutineDecrease(t *testing.T) {
	pool := NewPool(100, 1*time.Second)

	for i := 0; i < 10; i++ {
		t.Log(len(pool.workers))
		time.Sleep(1 * time.Second)
	}
	pool.WaitClose()

}
