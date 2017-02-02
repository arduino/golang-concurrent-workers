package cc

import "sync"

// Pool manages a pool of concurrent workers. It works a bit like a Waitgroup, but with error reporting and concurrency limits
// You create one with New, and run functions with Run. Then you wait on it like a regular WaitGroup and loop over the errors.
// It's important to loop over the errors because that's what's blocking.
//
// Example:
//
//   p := cc.New(4)
//   p.Run(func() {
//       p.Errors <- afunction()
//   })
//   p.Wait()
//
//   for err := range p.Errors {
//
//   }
type Pool struct {
	Errors chan error

	semaphore chan bool
	wg        *sync.WaitGroup
}

// New returns a new pool where a limited number (concurrency) of goroutine can work at the same time
func New(concurrency int) *Pool {
	wg := sync.WaitGroup{}
	p := Pool{
		Errors: make(chan error),

		semaphore: make(chan bool, concurrency),
		wg:        &wg,
	}
	return &p
}

// Wait doesn't block, but ensures that the channels are closed when all the goroutines end.
func (p *Pool) Wait() {
	go func() {
		p.wg.Wait()
		close(p.Errors)
		close(p.semaphore)
	}()
}

// Run wraps the given function into a goroutine and ensure that the concurrency limits are respected.
func (p *Pool) Run(fn func()) {
	p.wg.Add(1)
	go func() {
		p.semaphore <- true
		defer func() {
			<-p.semaphore
			p.wg.Done()
		}()
		fn()
	}()
}
