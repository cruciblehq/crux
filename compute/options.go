package compute

import (
	"io"

	"github.com/cruciblehq/crux/affordance/kernel"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Resource requirements for provisioning a compute host.
//
// Each backend maps these requirements to its native compute class. The local
// backend uses fixed resource allocation and ignores the sizing fields.
type Options struct {
	CPUs   int         // Minimum virtual CPUs required; zero means no minimum.
	Memory int         // Minimum memory in GiB required; zero means no minimum.
	Disk   int         // Minimum disk size in GiB required; zero means no minimum.
	Kernel kernel.Spec // Kernel requirements applied at provisioning time.
}

// Controls how a container process is executed.
//
// Each field configures one aspect of the process environment. Fields with
// zero values fall back to the corresponding defaults from the OCI spec or
// image configuration.
type RuntimeOptions struct {
	OCI     specs.Spec        // OCI runtime spec applied as the base configuration at container start.
	Shell   string            // Shell for "shell -c command" invocation; defaults to /bin/sh.
	Env     map[string]string // Environment variables merged into the spec env.
	Workdir string            // Working directory:
	User    string            // User identity as numeric "uid" or "uid:gid".
	Stdout  io.Writer         // Destination for process stdout.
	Stderr  io.Writer         // Destination for process stderr.
}
