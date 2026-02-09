//go:build !darwin && !linux

package runtime

import (
	"github.com/cruciblehq/crux/resource"
)

// Returns [ErrUnsupportedPlatform] on unsupported platforms.
func Start() error {
	return ErrUnsupportedPlatform
}

// Returns [ErrUnsupportedPlatform] on unsupported platforms.
func Stop() error {
	return ErrUnsupportedPlatform
}

// Returns [ErrUnsupportedPlatform] on unsupported platforms.
func Destroy() error {
	return ErrUnsupportedPlatform
}

// Returns [ErrUnsupportedPlatform] on unsupported platforms.
func GetStatus() (Status, error) {
	return StatusNotCreated, ErrUnsupportedPlatform
}

// Returns [ErrUnsupportedPlatform] on unsupported platforms.
func Exec(command string, args ...string) (*ExecResult, error) {
	return nil, ErrUnsupportedPlatform
}

// Returns [ErrUnsupportedPlatform] on unsupported platforms.
func ImportImage(_ string, _ resource.Type, _, _ string) error {
	return ErrUnsupportedPlatform
}
