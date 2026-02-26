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

// Restarts the container runtime environment.
//
// On macOS this stops and reboots the Lima virtual machine, preserving
// disk state. Blocks until the VM passes its readiness probes.
func Restart() error {
	l, err := newLima()
	if err != nil {
		return err
	}

	status, err := l.status()
	if err != nil {
		return err
	}

	if status == StateRunning {
		if err := l.stop(); err != nil {
			return err
		}
	}

	return l.start()
}

// Destroys and recreates the container runtime environment from scratch.
//
// On macOS this deletes the Lima virtual machine and all its data, then
// provisions a new one. Blocks until the VM passes its readiness probes.
func Reset() error {
	l, err := newLima()
	if err != nil {
		return err
	}

	status, err := l.status()
	if err != nil {
		return err
	}

	if status != StateNotCreated {
		if err := l.destroy(); err != nil {
			return err
		}
	}

	return l.start()
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
