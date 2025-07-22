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
	t.Run("Start", testStartContainer)
	t.Run("Stop", testStopContainer)
	t.Run("Remove", testRemoveContainer)
}

func testCreateContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	userID := 1
	teamID := 99
	imageName := "alpine"
	challengeName := "integration"
	containerID, err := docker.CreateContainer(ctx, userID, teamID, imageName, challengeName)
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
	err := docker.StartContainer(ctx, sharedContainerID, time.Now().Add(1*time.Minute))
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
