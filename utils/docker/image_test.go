package docker_test

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/intraware/rodan/utils/docker"
)

const testImage = "alpine:latest"

func TestImageExists_NotPulled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("Failed to create Docker client: %v", err)
	}
	cli.ImageRemove(ctx, testImage, image.RemoveOptions{Force: true})
	t.Logf("Removed existing %s image from the local repository", testImage)
	exists := docker.ImageExists(ctx, testImage)
	if exists {
		t.Error("Expected image not to exist, but it does")
	}
	t.Logf("Image %s does not exist. Test case passed", testImage)
}

func TestPullImage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := docker.PullImage(ctx, testImage)
	if err != nil {
		t.Fatalf("PullImage failed: %v", err)
	}
	t.Logf("Pulled image %s successfully", testImage)
	exists := docker.ImageExists(ctx, testImage)
	if !exists {
		t.Error("Expected image to exist after pull, but it does not")
	}
	t.Logf("Pulled image %s exists in the local repository. Test case passed", testImage)
}

func TestImageExists_Pulled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := docker.PullImage(ctx, testImage)
	if err != nil {
		t.Fatalf("Failed to pull image: %v", err)
	}
	t.Logf("Image %s pulled successfully", testImage)
	exists := docker.ImageExists(ctx, testImage)
	if !exists {
		t.Error("Expected image to exist, but it does not")
	}
	t.Logf("Image %s exists in the local repository. Test case passed", testImage)
}
