//go:build !darwin && !linux

package runtime

import (
	"context"
	"io"

	containerd "github.com/containerd/containerd/v2/client"
)

// Returns [ErrUnsupportedPlatform].
func containerCopy(_ context.Context, _ *containerd.Client, _, _ string, _ io.Reader, _ string) error {
	return ErrUnsupportedPlatform
}
