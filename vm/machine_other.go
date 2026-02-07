//go:build !darwin

package vm

// Handle to the crux virtual machine.
type Machine struct{}

// Returns [ErrUnsupportedPlatform] on non-darwin platforms.
func NewMachine() (*Machine, error) {
	return nil, ErrUnsupportedPlatform
}

// Returns [ErrUnsupportedPlatform] on non-darwin platforms.
func (m *Machine) Start() error {
	return ErrUnsupportedPlatform
}

// Returns [ErrUnsupportedPlatform] on non-darwin platforms.
func (m *Machine) Stop() error {
	return ErrUnsupportedPlatform
}

// Returns [ErrUnsupportedPlatform] on non-darwin platforms.
func (m *Machine) Destroy() error {
	return ErrUnsupportedPlatform
}

// Returns [ErrUnsupportedPlatform] on non-darwin platforms.
func (m *Machine) Status() (Status, error) {
	return StatusNotCreated, ErrUnsupportedPlatform
}

// Returns [ErrUnsupportedPlatform] on non-darwin platforms.
func (m *Machine) Exec(command string, args ...string) (*ExecResult, error) {
	return nil, ErrUnsupportedPlatform
}
