//go:build linux

package runtime

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/paths"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
)

// Returns the path to the ctr binary bundled with containerd.
func ctrPath() string {
	return filepath.Join(paths.Data(), containerdDir, "bin", "ctr")
}

// Returns the containerd socket address.
func containerdAddress() string {
	return filepath.Join(paths.Runtime(), containerdSock)
}

// Starts the container runtime environment.
func Start() error {
	return ErrUnsupportedPlatform
}

// Stops the container runtime environment.
func Stop() error {
	return ErrUnsupportedPlatform
}

// Destroys the container runtime environment and its resources.
func Destroy() error {
	return ErrUnsupportedPlatform
}

// Queries the current state of the container runtime environment.
func GetStatus() (Status, error) {
	return StatusNotCreated, ErrUnsupportedPlatform
}

// Runs a command inside the container runtime environment.
func Exec(command string, args ...string) (*ExecResult, error) {
	return nil, ErrUnsupportedPlatform
}

// Imports an OCI image tarball into containerd.
//
// The registry component of the resource reference is used as the containerd
// namespace for isolation. Images are tagged as "namespace/name:version".
// On Linux, ctr is invoked directly against the local containerd socket.
func ImportImage(ref string, typ resource.Type, version, path string) error {
	if _, err := os.Stat(path); err != nil {
		return crex.Wrap(ErrImageFileOpen, err)
	}

	id, err := reference.ParseIdentifier(ref, typ, nil)
	if err != nil {
		return crex.Wrap(ErrResourceRef, err)
	}

	ns := id.Registry()
	tag := id.Namespace() + "/" + id.Name() + ":" + version

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(
		ctrPath(),
		"--address", containerdAddress(),
		"--namespace", ns,
		"image", "import",
		"--tag", tag,
		path,
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return crex.Wrap(ErrImageImport, fmt.Errorf("ctr exited with code %d: %s", exitErr.ExitCode(), stderr.String()))
		}
		return crex.Wrap(ErrImageImport, err)
	}

	return nil
}
