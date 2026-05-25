package ctr

import "errors"

var (
	// Returned when the connection to the containerd socket fails.
	ErrConnect = errors.New("containerd connection failed")

	// Returned when importing an OCI image archive fails.
	ErrImport = errors.New("image import failed")

	// Returned when an import produces no images.
	ErrNoImages = errors.New("no images in archive")

	// Returned when creating or running a container fails.
	ErrRun = errors.New("container run failed")

	// Returned when a build step or image manipulation operation fails.
	ErrBuild = errors.New("build failed")
)
