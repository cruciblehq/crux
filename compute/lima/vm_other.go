//go:build !darwin && !linux

package lima

import (
	"context"
	"io"
	"net/url"

	"github.com/cruciblehq/utils-go/crex"
)

// System error returned by every operation on platforms Lima does not support.
var errUnsupportedPlatform = crex.SystemError("unsupported platform", "the local provider is not supported on this platform").
	Recovery("Run crux on macOS or Linux to use the local provider.").
	Cause(ErrUnsupportedPlatform).
	Err()

// Fails because the local provider does not support this platform.
func EnsureMachineImage(_ context.Context, _ *url.URL) (string, error) {
	return "", errUnsupportedPlatform
}

// Fails because the local provider does not support this platform.
func UploadImage(_ context.Context, _ string) (string, error) {
	return "", errUnsupportedPlatform
}

// Fails because the local provider does not support this platform.
func Provision(_ context.Context, _ string, _ []byte) error {
	return errUnsupportedPlatform
}

// Fails because the local provider does not support this platform.
func Deprovision(_ context.Context) error {
	return errUnsupportedPlatform
}

// Fails because the local provider does not support this platform.
func Start(_ context.Context) error {
	return errUnsupportedPlatform
}

// Fails because the local provider does not support this platform.
func Stop(_ context.Context) error {
	return errUnsupportedPlatform
}

// Fails because the local provider does not support this platform.
func Status(_ context.Context) (State, error) {
	return 0, errUnsupportedPlatform
}

// Fails because the local provider does not support this platform.
func Exec(_ context.Context, _, _ io.Writer, _ string, _ ...string) (int, error) {
	return -1, errUnsupportedPlatform
}

// Fails because the local provider does not support this platform.
func List(_ context.Context) ([]string, error) {
	return nil, errUnsupportedPlatform
}

// Returns an empty address because the local provider does not support this platform.
func CruxdAddress() string {
	return ""
}
