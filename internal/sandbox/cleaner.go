package sandbox

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/intraware/rodan/utils/docker"
	"github.com/intraware/rodan/utils/values"
)

type cleaner struct {
	BoxList       *list.List
	CleanInterval time.Time
	mu            sync.RWMutex
	wakeUp        chan struct{}
}

func newCleaner() *cleaner {
	cl := &cleaner{
		BoxList:       list.New(),
		CleanInterval: time.Time{},
		wakeUp:        make(chan struct{}, 1),
	}
	go cl.clean()
	if values.GetConfig().Docker.CleanOrphaned {
		go cl.clean_orphan()
	}
	return cl
}

func (c *cleaner) Add(box *SandBox) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.BoxList.PushBack(box)
	deadline, ok := box.Context.Deadline()
	if !ok {
		return
	}
	if c.CleanInterval.IsZero() || deadline.Before(c.CleanInterval) {
		c.CleanInterval = deadline
		select {
		case c.wakeUp <- struct{}{}:
		default:
		}
	}
}

func (c *cleaner) clean_orphan() {
	ctx := context.Background()
	for {
		containerList, err := docker.ListContainers(ctx)
		if err != nil {
			time.Sleep(30 * time.Second)
			continue
		}
		for _, ctr := range containerList {
			existsInPool := containerPool.CheckIfExists(ctr.ID)
			existsInCleaner := c.checkIfExists(ctr.ID)
			if existsInPool || existsInCleaner {
				continue
			}
			docker.StopContainer(ctx, ctr.ID)
		}
		time.Sleep(1 * time.Minute)
	}
}

func (c *cleaner) clean() {
	for {
		var nextExpiry time.Time
		if c.boxLength() == 0 {
			select {
			case <-time.After(1 * time.Second):
			case <-c.wakeUp:
			}
			continue
		}
		c.mu.Lock()
		for e := c.BoxList.Front(); e != nil; {
			next := e.Next()
			box := e.Value.(*SandBox)
			if box.Context.Err() != nil {
				box.Stop()
				c.BoxList.Remove(e)
			} else {
				if deadline, ok := box.Context.Deadline(); ok {
					if nextExpiry.IsZero() || deadline.Before(nextExpiry) {
						nextExpiry = deadline
					}
				}
			}
			e = next
		}
		c.CleanInterval = nextExpiry
		c.mu.Unlock()
		if nextExpiry.IsZero() {
			select {
			case <-time.After(1 * time.Second):
			case <-c.wakeUp:
			}
		} else {
			sleepDuration := time.Until(nextExpiry)
			if sleepDuration > 0 {
				select {
				case <-time.After(sleepDuration):
				case <-c.wakeUp:
				}
			}
		}
	}
}

func (c *cleaner) Remove(target *SandBox) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for e := c.BoxList.Front(); e != nil; e = e.Next() {
		box := e.Value.(*SandBox)
		if box == target {
			c.BoxList.Remove(e)
			break
		}
	}
	var next time.Time
	for e := c.BoxList.Front(); e != nil; e = e.Next() {
		box := e.Value.(*SandBox)
		if deadline, ok := box.Context.Deadline(); ok {
			if next.IsZero() || deadline.Before(next) {
				next = deadline
			}
		}
	}
	c.CleanInterval = next
	select {
	case c.wakeUp <- struct{}{}:
	default:
	}
}

func (c *cleaner) boxLength() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.BoxList.Len()
}

func (c *cleaner) checkIfExists(containerID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for e := c.BoxList.Front(); e != nil; e = e.Next() {
		box := e.Value.(*SandBox)
		if box.Container != nil && box.Container.ID == containerID {
			return true
		}
	}
	return false
}
