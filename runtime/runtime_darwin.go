//go:build darwin

package runtime

// Starts the container runtime environment.
//
// On macOS this creates and boots a Lima virtual machine. Blocks until the
// VM passes its readiness probes.
func Start() error {
	l, err := newLima()
	if err != nil {
		return err
	}
	return l.start()
}

// Stops the container runtime environment.
func Stop() error {
	l, err := newLima()
	if err != nil {
		return err
	}
	return l.stop()
}

// Destroys the container runtime environment and its resources.
func Destroy() error {
	l, err := newLima()
	if err != nil {
		return err
	}
	return l.destroy()
}

// Queries the current state of the container runtime environment.
func Status() (State, error) {
	l, err := newLima()
	if err != nil {
		return StateNotCreated, err
	}
	return l.status()
}

// Runs a command inside the container runtime environment.
//
// On macOS the command is executed inside the Lima virtual machine, blocking
// until the command completes.
func Exec(command string, args ...string) (*ExecResult, error) {
	l, err := newLima()
	if err != nil {
		return nil, err
	}
	return l.exec(command, args...)
}
