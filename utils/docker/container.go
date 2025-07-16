package docker

import "time"

func StartContainer(containerID string, ttl time.Time) {
	// cli, err := getDockerClient()
}

func StopContainer(containerID string)                         {}
func CreateContainer(imageName string)                         {}
func RemoveContainer(containerID string)                       {}
func ExtendContainer(containerID string, extendTime time.Time) {}
