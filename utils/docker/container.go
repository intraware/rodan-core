package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
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

func CreateContainer(ctx context.Context, containerName, imageName string) (containerID string, err error) {
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
		Labels: map[string]string{
			"created_by": "rodan",
		},
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

func ListContainers(ctx context.Context) ([]container.Summary, error) {
	cli, err := GetDockerClient()
	if _, ping_err := cli.Ping(ctx); ping_err != nil {
		ResetDockerClient()
	}
	if err != nil {
		return nil, err
	}
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", "created_by=rodan")
	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, err
	}
	return containers, nil
}
