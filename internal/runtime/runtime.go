//go:build !darwin && !linux

package runtime

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
func Status() (State, error) {
	return StateNotCreated, ErrUnsupportedPlatform
}

// Returns [ErrUnsupportedPlatform] on unsupported platforms.
func Exec(command string, args ...string) (*ExecResult, error) {
	return nil, ErrUnsupportedPlatform
}
