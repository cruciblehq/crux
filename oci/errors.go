package oci

import "errors"

var (
	ErrInvalidImage          = errors.New("invalid OCI image")
	ErrInvalidPlatform       = errors.New("invalid platform format, expected os/arch")
	ErrInsufficientPlatforms = errors.New("image missing required platforms")
	ErrLayoutWrite           = errors.New("failed to write OCI layout")
	ErrLayerCreate           = errors.New("failed to create image layer")
	ErrImageBuild            = errors.New("failed to build image")
)
