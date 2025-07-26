package sandbox

import (
	"context"
	"time"

	"github.com/intraware/rodan/utils/docker"
)

type container struct {
	Context     context.Context
	ContainerID string
	ImageName   string
	ChallengeID int
	TTL         time.Duration
	StartedAt   time.Time
}

func newContainer(ctx context.Context, challengeID int, containerName, imageName string, ttl time.Duration) (*container, error) {
	if !docker.ImageExists(ctx, imageName) {
		return nil, errImageNotExists
	}
	containerID, err := docker.CreateContainer(ctx, containerName, imageName)
	if err != nil {
		return nil, err
	}
	return &container{
		Context:     ctx,
		ContainerID: containerID,
		ImageName:   imageName,
		ChallengeID: challengeID,
		TTL:         ttl,
	}, nil
}

func (c *container) Start() (err error) {
	err = docker.StartContainer(c.Context, c.ContainerID)
	if err == nil {
		c.StartedAt = time.Now()
	}
	return
}

func (c *container) Stop() (err error) {
	err = docker.StopContainer(c.Context, c.ContainerID)
	return
}

func (c *container) Discard() (err error) {
	err = docker.RemoveContainer(c.Context, c.ContainerID)
	return
}

func (c *container) Reset() {
	// run the cleaner script
}

func (c *container) GenerateFlag(flag string) {
	// run the generator script
}
