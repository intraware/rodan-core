package docker

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/go-connections/nat"
	"github.com/intraware/rodan/internal/utils/values"
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

func randomPortInRange(minPort, maxPort int) (int, error) {
	if minPort >= maxPort {
		return 0, errors.New("invalid port range")
	}
	portRange := maxPort - minPort + 1
	attempts := portRange * values.GetConfig().Docker.PortsMaxRetry
	for range attempts {
		port := rand.Intn(portRange) + minPort
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, errors.New("no available port found in range")
}

func CreateContainer(ctx context.Context, containerName, imageName string, internalPorts []string) (containerID string, err error) {
	containerID = ""
	cli, err := GetDockerClient()
	if err != nil {
		return
	}
	if _, pingErr := cli.Ping(ctx); pingErr != nil {
		ResetDockerClient()
	}
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	minPort := values.GetConfig().Docker.PortRange.Start
	maxPort := values.GetConfig().Docker.PortRange.End
	usedHostPorts := make(map[int]bool)
	for _, internal := range internalPorts {
		containerPort := nat.Port(internal + "/tcp")
		var hostPort int
		for {
			hostPort, err = randomPortInRange(minPort, maxPort)
			if err != nil {
				return
			}
			if !usedHostPorts[hostPort] {
				usedHostPorts[hostPort] = true
				break
			}
		}
		portBindings[containerPort] = []nat.PortBinding{{
			HostIP:   values.GetConfig().Docker.BindingHost,
			HostPort: fmt.Sprintf("%d", hostPort),
		}}
		exposedPorts[containerPort] = struct{}{}
	}
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		ExposedPorts: exposedPorts,
		Labels: map[string]string{
			"created_by": "rodan",
		},
	}, &container.HostConfig{
		PortBindings: portBindings,
	}, nil, nil, containerName)
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

func RunCommand(ctx context.Context, containerID, command string) (err error) {
	err = nil
	cli, err := GetDockerClient()
	if _, ping_err := cli.Ping(ctx); ping_err != nil {
		ResetDockerClient()
	}
	if err != nil {
		return
	}
	cmd := strings.Fields(command)
	if len(cmd) == 0 {
		return fmt.Errorf("empty command")
	}
	exe, err := cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		User:       "rodan",
		Privileged: false,
		Tty:        false,
		Cmd:        cmd,
		WorkingDir: "/",
	})
	if err != nil {
		return
	}
	err = cli.ContainerExecStart(ctx, exe.ID, container.ExecStartOptions{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return
	}
	return
}

func GetBoundPorts(ctx context.Context, containerID string) (map[string]string, error) {
	cli, err := GetDockerClient()
	if err != nil {
		return nil, err
	}
	info, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, err
	}
	bindings := make(map[string]string)
	for port, binding := range info.NetworkSettings.Ports {
		if len(binding) > 0 {
			bindings[string(port)] = binding[0].HostPort
		}
	}
	return bindings, nil
}
