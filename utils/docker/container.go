package docker

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/intraware/rodan/utils/values"
)

func StartContainer(ctx context.Context, containerID string) (err error) {
	err = nil
	cli, err := GetDockerClient()
	if _, ping_err := cli.Ping(ctx); ping_err != nil {
		ResetDockerClient()
	}
	if err != nil {
		return
	}
	err = cli.ContainerStart(ctx, containerID, container.StartOptions{})
	return
}

func StopContainer(ctx context.Context, containerID string) (err error) {
	err = nil
	cli, err := GetDockerClient()
	if _, ping_err := cli.Ping(ctx); ping_err != nil {
		ResetDockerClient()
	}
	if err != nil {
		return
	}
	timeout := int(values.GetConfig().Docker.ContainerTimeout.Seconds())
	err = cli.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	})
	return
}

func CreateContainer(ctx context.Context, containerName, imageName string, ttl time.Time) (containerID string, err error) {
	containerID = ""
	err = nil
	cli, err := GetDockerClient()
	if _, ping_err := cli.Ping(ctx); ping_err != nil {
		ResetDockerClient()
	}
	if err != nil {
		return
	}
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
	}, &container.HostConfig{}, nil, nil, containerName)
	if err != nil {
		return
	}
	containerID = resp.ID
	return
}

func RemoveContainer(ctx context.Context, containerID string) (err error) {
	err = nil
	cli, err := GetDockerClient()
	if _, ping_err := cli.Ping(ctx); ping_err != nil {
		ResetDockerClient()
	}
	if err != nil {
		return
	}
	err = cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	})
	return
}
