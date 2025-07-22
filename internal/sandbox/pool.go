package sandbox

import "sync"

type pool struct {
	mu   sync.RWMutex
	pool map[int]*[]container
}

func newPool() *pool {
	return &pool{
		pool: make(map[int]*[]container),
	}
}

func (p *pool) Aquire(challengeID int) (*container, error) {
	return nil, nil
}

func (p *pool) Release(container *container) error {
	return nil
}
