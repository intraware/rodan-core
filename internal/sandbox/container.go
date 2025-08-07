package sandbox

import (
	"context"
	"fmt"
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

func newContainer(ctx context.Context, challengeID int, containerName, imageName string, ttl time.Duration, exposedPorts []string) (*container, error) {
	if !docker.ImageExists(ctx, imageName) {
		return nil, errImageNotExists
	}
	containerID, err := docker.CreateContainer(ctx, containerName, imageName, exposedPorts)
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

func (c *container) Reset() (err error) {
	err = docker.RunCommand(c.Context, c.ContainerID, "./reset")
	return
}

func (c *container) GenerateFlag(flag string) (err error) {
	generate := fmt.Sprintf("./generate %s", flag)
	err = docker.RunCommand(c.Context, c.ContainerID, generate)
	return
}

func (c *container) GetAll() ([]string, error) {
	containers, err := docker.ListContainers(c.Context)
	if err != nil {
		return nil, err
	}
	var containerIDs []string
	for _, ctr := range containers {
		containerIDs = append(containerIDs, ctr.ID)
	}
	return containerIDs, nil
}
