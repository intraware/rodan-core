package docker_test

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/intraware/rodan/utils/docker"
)

var sharedContainerID string

func TestContainerLifecycleSequence(t *testing.T) {
	t.Run("Create", testCreateContainer)
	t.Run("List", testListContainers)
	t.Run("Start", testStartContainer)
	t.Run("RunCommand", testRunCommand)
	t.Run("Stop", testStopContainer)
	t.Run("Remove", testRemoveContainer)
}

func testCreateContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	imageName := "alpine"
	containerName := "integration-test"
	containerID, err := docker.CreateContainer(ctx, containerName, imageName, nil)
	if err != nil {
		t.Fatalf("CreateContainer failed: %v", err)
	}
	if containerID == "" {
		t.Fatal("CreateContainer returned empty ID")
	}
	sharedContainerID = containerID
	t.Logf("Container created: %s", sharedContainerID)
}

func testStartContainer(t *testing.T) {
	if sharedContainerID == "" {
		t.Fatal("sharedContainerID is empty — Create test probably failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := docker.StartContainer(ctx, sharedContainerID)
	if err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}
	t.Logf("Container started: %s", sharedContainerID)
}

func testStopContainer(t *testing.T) {
	if sharedContainerID == "" {
		t.Fatal("sharedContainerID is empty — Start test probably failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := docker.StopContainer(ctx, sharedContainerID)
	if err != nil {
		t.Fatalf("StopContainer failed: %v", err)
	}
	t.Logf("Container stopped: %s", sharedContainerID)
}

func testRemoveContainer(t *testing.T) {
	if sharedContainerID == "" {
		t.Fatal("sharedContainerID is empty — Stop test probably failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := docker.RemoveContainer(ctx, sharedContainerID)
	if err != nil {
		t.Fatalf("RemoveContainer failed: %v", err)
	}
	t.Logf("Container removed: %s", sharedContainerID)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("Failed to create Docker client: %v", err)
	}
	_, err = cli.ContainerInspect(ctx, sharedContainerID)
	if err == nil {
		t.Error("Expected container to be removed, but it's still inspectable")
	} else {
		t.Log("Container removal verified")
	}
}

func testRunCommand(t *testing.T) {
	if sharedContainerID == "" {
		t.Fatal("sharedContainerID is empty — Start test probably failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := docker.RunCommand(ctx, sharedContainerID, "echo hello")
	if err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}
	t.Logf("Command ran successfully on container: %s", sharedContainerID)
}

func testListContainers(t *testing.T) {
	if sharedContainerID == "" {
		t.Fatal("sharedContainerID is empty — Create test probably failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	containers, err := docker.ListContainers(ctx)
	if err != nil {
		t.Fatalf("ListContainers failed: %v", err)
	}
	if len(containers) == 0 {
		t.Fatal("No containers returned by ListContainers")
	}
	found := false
	for _, c := range containers {
		if c.ID == sharedContainerID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected container %s not found in ListContainers result", sharedContainerID)
	} else {
		t.Logf("Container %s found in ListContainers", sharedContainerID)
	}
}
