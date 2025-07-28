package sandbox

import (
	"sync"

	"github.com/intraware/rodan/utils/values"
)

type pool struct {
	mu   sync.RWMutex
	pool map[int][]*container
}

func newPool() *pool {
	return &pool{
		pool: make(map[int][]*container),
	}
}

func (p *pool) Aquire(challengeID int) (*container, error) {
	var ctr *container
	p.mu.Lock()
	defer p.mu.Unlock()
	containers, ok := p.pool[challengeID]
	if !ok || len(containers) == 0 {
		return nil, errNoContainers
	}
	ctr = containers[0]
	p.pool[challengeID] = containers[1:]
	if len(p.pool[challengeID]) == 0 {
		delete(p.pool, challengeID)
	}
	return ctr, nil
}

func (p *pool) Release(c *container) error {
	challengeID := c.ChallengeID
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.pool[challengeID]) >= values.GetConfig().Docker.PoolSize {
		return errPoolFull
	}
	if len(p.pool[challengeID]) == 0 {
		p.pool[challengeID] = make([]*container, values.GetConfig().Docker.PoolSize)
	}
	p.pool[challengeID] = append(p.pool[challengeID], c)
	return nil
}

func (p *pool) CheckIfExists(containerID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, containers := range p.pool {
		for i := range containers {
			ctr := containers[i]
			if ctr.ContainerID == containerID {
				return true
			}
		}
	}
	return false
}
