package workerpool

import (
	"fmt"
)

type Worker struct {
	pool *Pool
	task chan func()
}

func (w *Worker) run() {
	go func() {
		defer func() {
			if p := recover(); p != nil {
				fmt.Printf("worker exits from a panic: %v\n", p)
				w.run()
				w.pool.revertWorker(w)
			}
		}()

		for f := range w.task {
			if f == nil {
				w.pool.workerCache.Put(w)
				return
			}
			f()
			w.pool.revertWorker(w)
		}
	}()
}
