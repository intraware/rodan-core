package sandbox

import "errors"

var errImageNotExists = errors.New("Image does not exist")
var errNoContainers = errors.New("no containers exist")
var errPoolFull = errors.New("container pool is full")

var ErrFailedToCreateContainer = errors.New("Failed to create a new container")
var ErrContainerNotFound = errors.New("Container not found")
var ErrFailedToDiscardContainer = errors.New("Failed to discard container")
var ErrFailedToStartContainer = errors.New("Failed to start the created container")
var ErrFailedToStopContainer = errors.New("Failed to stop the container")
