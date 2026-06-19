//go:build !darwin && !linux

package local

import (
	"context"
	"io"

	"github.com/cruciblehq/spec/affordance/kernel"
	"github.com/cruciblehq/utils-go/crex"
)

// System error returned by every local compute operation on platforms where the
// backend is unavailable.
var errUnsupportedPlatform = crex.SystemError("unsupported platform", "the local compute backend is not supported on this platform").
	Recovery("Run crux on macOS or Linux to use the local backend.").
	Cause(ErrUnsupportedPlatform).
	Err()

func ensureMachineImage(_ context.Context) (string, error) {
	return "", errUnsupportedPlatform
}

func uploadImage(_ context.Context, _ string) (string, error) {
	return "", errUnsupportedPlatform
}

func provision(_ context.Context, _, _ string, _ kernel.Spec) error {
	return errUnsupportedPlatform
}

func deprovision(_ context.Context, _ string) error {
	return errUnsupportedPlatform
}

func start(_ context.Context, _ string) error {
	return errUnsupportedPlatform
}

func stop(_ context.Context, _ string) error {
	return errUnsupportedPlatform
}

func status(_ context.Context, _ string) (State, error) {
	return 0, errUnsupportedPlatform
}

func execute(_ context.Context, _ string, _ io.Writer, _ io.Writer, _ string, _ ...string) (int, error) {
	return -1, errUnsupportedPlatform
}

func list(_ context.Context) ([]string, error) {
	return nil, errUnsupportedPlatform
}

func copyArchive(_ context.Context, _ string, _ io.Reader) error {
	return errUnsupportedPlatform
}

func containerdSocket(_ context.Context, _ string) (string, error) {
	return "", errUnsupportedPlatform
}
