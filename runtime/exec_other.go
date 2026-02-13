//go:build !darwin && !linux

package runtime

import "context"

// Returns [ErrUnsupportedPlatform].
func containerExec(_ context.Context, _, _ string, _ ExecOptions, _ string, _ ...string) (*ExecResult, error) {
	return nil, ErrUnsupportedPlatform
}
