package docker

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/docker/docker/client"
)

var (
	dockerClient atomic.Value
	initMutex    sync.Mutex
)

func getDockerClient() (*client.Client, error) {
	if cli := dockerClient.Load(); cli != nil {
		return cli.(*client.Client), nil
	}
	initMutex.Lock()
	defer initMutex.Unlock()
	if cli := dockerClient.Load(); cli != nil {
		return cli.(*client.Client), nil
	}
	//	cfg := values.GetConfig()
	socketURL := "unix:///var/run/docker.sock"
	var httpClient *http.Client
	switch {
	case strings.HasPrefix(socketURL, "unix://"):
		httpClient = &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", socketURL[len("unix://"):])
				},
			},
		}
	case strings.HasPrefix(socketURL, "tcp://"):
		httpClient = http.DefaultClient
	default:
		return nil, fmt.Errorf("unsupported socket URL scheme: %s", socketURL)
	}
	cli, err := client.NewClientWithOpts(
		client.WithHost(socketURL),
		client.WithHTTPClient(httpClient),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	dockerClient.Store(cli)
	return cli, nil
}

func resetDockerClient() {
	dockerClient.Store(nil)
}
