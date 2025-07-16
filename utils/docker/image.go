package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/image"
)

func ImageExists(ctx context.Context, imageName string) bool {
	cli, err := getDockerClient()
	if err != nil {
		return false
	}
	_, err = cli.ImageInspect(ctx, imageName)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false
		}
		return false
	}
	return true
}

func PullImage(ctx context.Context, imageName string) error {
	cli, err := getDockerClient()
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}
	reader, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()
	_, err = io.Copy(io.Discard, reader) // for now ... later will add to logs support
	if err != nil {
		return fmt.Errorf("failed to read image pull response: %w", err)
	}
	return nil
}
