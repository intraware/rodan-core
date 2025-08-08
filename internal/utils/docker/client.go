package docker

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/intraware/rodan/internal/utils/values"
)

var (
	dockerClient *client.Client
)

func SetupDockerClient() (err error) {
	cfg := values.GetConfig().Docker
	socketURL := cfg.SocketURL
	if socketURL == "" {
		err = fmt.Errorf("docker socket URL is not configured")
		return
	}
	dockerClient, err = client.NewClientWithOpts(
		client.WithHost(socketURL),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		err = fmt.Errorf("failed to create docker client: %w", err)
		return
	}
	return
}

// immutable getter for dockerClient
func GetDockerClient() *client.Client {
	return dockerClient
}
