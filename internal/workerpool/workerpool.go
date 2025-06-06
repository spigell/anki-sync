package workerpool

import (
	"context"
	"sync"
)

type TaskFunc func(ctx context.Context) error

type Pool struct {
	parallelism int
	tasks       chan TaskFunc
	wg          sync.WaitGroup
}

func New(parallelism int) *Pool {
	return &Pool{
		parallelism: parallelism,
		tasks:       make(chan TaskFunc),
	}
}

func (p *Pool) Start(ctx context.Context) {
	for range p.parallelism {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-p.tasks:
					if !ok {
						return
					}
					_ = task(ctx) // log or handle error if needed
				}
			}
		}()
	}
}

func (p *Pool) Submit(task TaskFunc) {
	p.tasks <- task
}

func (p *Pool) Stop() {
	close(p.tasks)
	p.wg.Wait()
}
