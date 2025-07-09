// filepath: /storage/Projects/rodan/utils/docker.go
package utils

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils/values"
	"github.com/sirupsen/logrus"
)

type DockerService struct {
	client     *client.Client
	portRanges []PortRange
}

type PortRange struct {
	Start int
	End   int
}

type ContainerInfo struct {
	ID    string
	Ports []int
	Links []string
}

// CleanupService handles cleanup of expired containers
type CleanupService struct {
	dockerService *DockerService
}

func NewDockerService() (*DockerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}
	cfg := values.GetConfig()
	// Parse port ranges from config
	portRanges := []PortRange{
		{Start: cfg.Docker.PortRange.Start, End: cfg.Docker.PortRange.End},
	}

	return &DockerService{
		client:     cli,
		portRanges: portRanges,
	}, nil
}

func (ds *DockerService) CreateContainer(challengeID, teamID int, imageName string, exposedPorts []string) (*ContainerInfo, error) {
	ctx := context.Background()

	// Generate container name
	containerName := fmt.Sprintf("challenge_%d_team_%d_%d", challengeID, teamID, time.Now().Unix())

	// Allocate random ports within range
	hostPorts, err := ds.allocateRandomPorts(len(exposedPorts))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate ports: %v", err)
	}

	// Create port binding
	portBindings := nat.PortMap{}
	exposedPortSet := nat.PortSet{}

	for i, port := range exposedPorts {
		containerPort := nat.Port(port)
		exposedPortSet[containerPort] = struct{}{}
		portBindings[containerPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: strconv.Itoa(hostPorts[i]),
			},
		}
	}

	// Container configuration
	containerConfig := &container.Config{
		Image:        imageName,
		ExposedPorts: exposedPortSet,
		Labels: map[string]string{
			"challenge_id": strconv.Itoa(challengeID),
			"team_id":      strconv.Itoa(teamID),
			"managed_by":   "rodan",
		},
	}

	// Host configuration
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		AutoRemove:   true,
		RestartPolicy: container.RestartPolicy{
			Name: "no",
		},
		// Security settings
		Privileged:     false,
		ReadonlyRootfs: false,
		// Resource limits
		Resources: container.Resources{
			Memory:   512 * 1024 * 1024, // 512MB
			CPUQuota: 50000,             // 50% CPU
		},
	}

	// Network configuration
	networkConfig := &network.NetworkingConfig{}

	// Create container
	resp, err := ds.client.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	// Start container
	if err := ds.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		// Clean up if start fails
		ds.client.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	// Generate links
	cfg := values.GetConfig()
	links := make([]string, len(hostPorts))
	for i, port := range hostPorts {
		links[i] = fmt.Sprintf("http://%s:%d", cfg.Server.Host, port)
	}

	logrus.Infof("Created container %s for challenge %d team %d", resp.ID[:12], challengeID, teamID)

	return &ContainerInfo{
		ID:    resp.ID,
		Ports: hostPorts,
		Links: links,
	}, nil
}

func (ds *DockerService) StopContainer(containerID string) error {
	ctx := context.Background()

	// Stop container with a timeout
	timeout := 30
	if err := ds.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		logrus.Warnf("Failed to stop container %s gracefully: %v", containerID[:12], err)
	}

	// Force remove container
	if err := ds.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("failed to remove container: %v", err)
	}

	logrus.Infof("Stopped and removed container %s", containerID[:12])
	return nil
}

func (ds *DockerService) GetContainerStatus(containerID string) (string, error) {
	ctx := context.Background()

	containerJSON, err := ds.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %v", err)
	}

	return containerJSON.State.Status, nil
}

func (ds *DockerService) CleanupExpiredContainers(maxAge time.Duration) error {
	ctx := context.Background()

	// List containers with our label
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", "managed_by=rodan")

	containers, err := ds.client.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %v", err)
	}

	for _, container := range containers {
		// Check if container is older than maxAge
		if time.Since(time.Unix(container.Created, 0)) > maxAge {
			logrus.Infof("Cleaning up expired container %s", container.ID[:12])
			if err := ds.StopContainer(container.ID); err != nil {
				logrus.Errorf("Failed to cleanup container %s: %v", container.ID[:12], err)
			}
		}
	}

	return nil
}

func (ds *DockerService) allocateRandomPorts(count int) ([]int, error) {
	if len(ds.portRanges) == 0 {
		return nil, fmt.Errorf("no port ranges configured")
	}

	// For simplicity, use the first port range
	portRange := ds.portRanges[0]
	availablePorts := portRange.End - portRange.Start + 1

	if count > availablePorts {
		return nil, fmt.Errorf("not enough ports available in range")
	}

	// Generate random ports within range
	rand.Seed(time.Now().UnixNano())
	usedPorts := make(map[int]bool)
	ports := make([]int, 0, count)

	for len(ports) < count {
		port := rand.Intn(availablePorts) + portRange.Start
		if !usedPorts[port] {
			// TODO: Check if port is actually available on the host
			usedPorts[port] = true
			ports = append(ports, port)
		}
	}

	return ports, nil
}

func (ds *DockerService) Close() error {
	return ds.client.Close()
}

// NewCleanupService creates a new cleanup service
func NewCleanupService() (*CleanupService, error) {
	dockerService, err := NewDockerService()
	if err != nil {
		return nil, err
	}

	return &CleanupService{
		dockerService: dockerService,
	}, nil
}

// StartCleanupRoutine starts a background routine to cleanup expired containers
func (cs *CleanupService) StartCleanupRoutine() {
	// Run cleanup every 30 minutes
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for range ticker.C {
			cs.CleanupExpiredContainers(2 * time.Hour) // Cleanup containers older than 2 hours
		}
	}()
}

// CleanupExpiredContainers removes containers that have been running for too long
func (cs *CleanupService) CleanupExpiredContainers(maxAge time.Duration) {
	logrus.Info("Starting container cleanup routine")

	// Get all containers from database
	var containers []models.Container
	if err := models.DB.Find(&containers).Error; err != nil {
		logrus.Errorf("Failed to fetch containers from database: %v", err)
		return
	}

	for _, container := range containers {
		// Check if container still exists in Docker
		status, err := cs.dockerService.GetContainerStatus(container.ContainerID)
		if err != nil {
			// Container doesn't exist in Docker, remove from database
			logrus.Warnf("Container %s not found in Docker, removing from database", container.ContainerID[:12])
			models.DB.Delete(&container)
			continue
		}

		// If container is not running, remove it
		if status != "running" {
			logrus.Infof("Container %s is not running (status: %s), cleaning up", container.ContainerID[:12], status)
			cs.dockerService.StopContainer(container.ContainerID)
			models.DB.Delete(&container)
		}
	}

	// Also cleanup containers directly from Docker that might not be in our database
	if err := cs.dockerService.CleanupExpiredContainers(maxAge); err != nil {
		logrus.Errorf("Failed to cleanup expired containers from Docker: %v", err)
	}

	logrus.Info("Container cleanup routine completed")
}

// Close closes the cleanup service
func (cs *CleanupService) Close() error {
	return cs.dockerService.Close()
}
