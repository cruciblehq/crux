//go:build !darwin && !linux

package local

import (
	"context"
	"io"

	"github.com/cruciblehq/crux/compute/provider"
	"github.com/cruciblehq/crux/manifest"
)

// Not supported on this platform.
func ensureMachineImage(_ context.Context) (string, error) {
	return "", ErrUnsupportedPlatform
}

// Not supported on this platform.
func uploadImage(_ context.Context, _ string) (string, error) {
	return "", ErrUnsupportedPlatform
}

// Not supported on this platform.
func provision(_ context.Context, _, _ string, _ *manifest.ComputePolicy) error {
	return ErrUnsupportedPlatform
}

// Not supported on this platform.
func deprovision(_ context.Context, _ string) error {
	return ErrUnsupportedPlatform
}

// Not supported on this platform.
func start(_ context.Context, _ string) error {
	return ErrUnsupportedPlatform
}

// Not supported on this platform.
func stop(_ context.Context, _ string) error {
	return ErrUnsupportedPlatform
}

// Not supported on this platform.
func status(_ context.Context, _ string) (provider.State, error) {
	return 0, ErrUnsupportedPlatform
}

// Not supported on this platform.
func execute(_ context.Context, _ string, _ io.Writer, _ io.Writer, _ string, _ ...string) (int, error) {
	return -1, ErrUnsupportedPlatform
}
