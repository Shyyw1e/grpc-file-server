package limiter


type Semafore chan struct{}

func NewSemafore(limit int) Semafore {
	return make(chan struct{}, limit)
}

func (s Semafore)Acquire() {
	s <-struct{}{}
}

func (s Semafore) Release() {
	<-s
}

