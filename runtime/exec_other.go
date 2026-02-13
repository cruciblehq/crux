//go:build !darwin && !linux

package runtime

import (
	"context"

	containerd "github.com/containerd/containerd/v2/client"
)

// Returns [ErrUnsupportedPlatform].
func containerExec(_ context.Context, _ *containerd.Client, _, _ string, _ ExecOptions, _ string, _ ...string) (*ExecResult, error) {
	return nil, ErrUnsupportedPlatform
}
