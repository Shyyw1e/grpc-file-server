package limiter

import "context"

type Semaphore chan struct{}

func NewSemaphore(limit int) Semaphore {
	return make(chan struct{}, limit)
}

func (s Semaphore) Acquire() {
	s <- struct{}{}
}

func (s Semaphore) Release() {
	<-s
}

func (s Semaphore) TryAcquire(ctx context.Context) bool {
	select {
	case s <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}
