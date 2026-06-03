package resource

import (
	"io"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Configuration for a single Run step executed inside a build container.
//
// The backend creates a container from imageRef, runs the command using Shell,
// commits the resulting filesystem diff as a new image layer, and removes the
// container before returning. Every field is optional; zero values select the
// backend's defaults (typically /bin/sh, the image's default working directory
// and user, and the process's stdout/stderr).
type Step struct {
	Shell    string      // Shell executable used to run the command (defaults to /bin/sh when empty).
	Command  string      // Command string passed to the shell as a single argument.
	Env      []string    // Environment variables for the step.
	Workdir  string      // Working directory override for the step.
	User     string      // User identity override for the step in "uid:gid" format.
	Security *specs.Spec // OCI runtime spec used to configure the build container's security policy.
	Stdout   io.Writer   // Destination for the command's standard output.
	Stderr   io.Writer   // Destination for the command's standard error.
}
