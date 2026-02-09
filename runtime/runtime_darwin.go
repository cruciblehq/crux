//go:build darwin

package runtime

import (
	"github.com/cruciblehq/crux/resource"
)

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
func GetStatus() (Status, error) {
	l, err := newLima()
	if err != nil {
		return StatusNotCreated, err
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

// Imports an OCI image tarball into containerd.
//
// The registry component of the resource reference is used as the containerd
// namespace for isolation. Images are tagged as "namespace/name:version". On
// macOS the import runs inside the Lima virtual machine.
func ImportImage(ref string, typ resource.Type, version, path string) error {
	l, err := newLima()
	if err != nil {
		return err
	}
	return l.importImage(ref, typ, version, path)
}
