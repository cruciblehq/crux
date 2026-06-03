//go:build !darwin && !linux

package local

import (
	"context"
	"io"

	"github.com/cruciblehq/crux/security/vm"
)

func ensureMachineImage(_ context.Context) (string, error) {
	return "", ErrUnsupportedPlatform
}

func uploadImage(_ context.Context, _ string) (string, error) {
	return "", ErrUnsupportedPlatform
}

func provision(_ context.Context, _, _ string, _ vm.VM) error {
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

func status(_ context.Context, _ string) (State, error) {
	return 0, ErrUnsupportedPlatform
}

func execute(_ context.Context, _ string, _ io.Writer, _ io.Writer, _ string, _ ...string) (int, error) {
	return -1, ErrUnsupportedPlatform
}

func list(_ context.Context) ([]string, error) {
	return nil, ErrUnsupportedPlatform
}

func copyArchive(_ context.Context, _ string, _ io.Reader) error {
	return ErrUnsupportedPlatform
}

func containerdSocket(_ context.Context, _ string) (string, error) {
	return "", ErrUnsupportedPlatform
}
