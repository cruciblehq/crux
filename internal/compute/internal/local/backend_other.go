//go:build !darwin && !linux

package local

import (
	"context"

	"github.com/cruciblehq/crux/internal/compute/internal/provider"
)

func provision(_ context.Context, _ *provider.Config) error {
	return ErrUnsupportedPlatform
}

func deprovision(_ context.Context, _ string) error {
	return ErrUnsupportedPlatform
}

func start(_ context.Context, _ string) error {
	return ErrUnsupportedPlatform
}

func stop(_ context.Context, _ string) error {
	return ErrUnsupportedPlatform
}

func status(_ context.Context, _ string) (provider.State, error) {
	return 0, ErrUnsupportedPlatform
}

func execute(_ context.Context, _ string, _ string, _ ...string) (*provider.ExecResult, error) {
	return nil, ErrUnsupportedPlatform
}

func newClient(_ string) (provider.Client, error) {
	return nil, ErrUnsupportedPlatform
}
