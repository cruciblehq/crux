package provider

import (
	"context"
	"io"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Low-level interface for building OCI images via containerd primitives.
//
// Each method that produces a new image returns an opaque image reference that
// can be passed to subsequent calls. References are scoped to the lifetime of
// the ImageBuilder and must not be used after Close is called. The builder
// owns all intermediate images it creates and removes them on Close.
type ImageBuilder interface {

	// Loads an OCI image archive into the builder's image store.
	//
	// The reader must contain a valid OCI image layout tar. If the archive
	// contains multiple manifests, only the first is used. Returns an opaque
	// image reference for use with other methods.
	Import(ctx context.Context, r io.Reader) (string, error)

	// Executes a shell command in a build container derived from imageRef.
	//
	// Creates a container from imageRef, runs the command, commits the
	// resulting filesystem diff as a new image layer, and removes the
	// container before returning. A non-zero exit code is an error.
	Run(ctx context.Context, imageRef string, cfg *RunConfig) (string, error)

	// Appends the uncompressed tar archive r as a new layer on top of imageRef.
	//
	// The implementation compresses the stream before storing it; callers must
	// not pre-compress r. Returns the updated image reference.
	Copy(ctx context.Context, imageRef string, r io.Reader) (string, error)

	// Applies configuration changes to imageRef without adding a new layer.
	//
	// Only non-zero fields in cfg are applied. Returns the updated image
	// reference with the new configuration.
	Configure(ctx context.Context, imageRef string, cfg *ConfigUpdate) (string, error)

	// Exports imageRef as an OCI image layout tar to w.
	//
	// The output format is compatible with Import. Only the specified image is
	// exported; intermediate images are not included.
	Export(ctx context.Context, imageRef string, w io.Writer) error

	// Extracts srcPath from imageRef as an uncompressed tar stream.
	//
	// When srcPath names a file the stream contains one entry. When it names a
	// directory the stream contains the directory and all its descendants. Layer
	// ordering follows overlayfs semantics: later layers override earlier ones;
	// whiteout entries delete files from earlier layers. Returns an error when
	// srcPath matches no entry in any layer. The caller is responsible for
	// closing the returned ReadCloser.
	Extract(ctx context.Context, imageRef, srcPath string) (io.ReadCloser, error)

	// Releases resources held by the builder.
	//
	// Removes intermediate images created during the build and closes the
	// connection to the containerd daemon. Must be called even when the build
	// fails.
	Close() error
}

// Configuration for a single command executed inside a build container.
//
// All fields are optional; zero values select the backend's defaults (typically
// /bin/sh for Shell, the image's default working directory and user, and the
// process's stdout/stderr for output).
type RunConfig struct {
	Shell    string      // Shell executable used to run Command (defaults to /bin/sh when empty).
	Command  string      // Command string passed to Shell as a single argument.
	Env      []string    // Environment variables for the build container.
	WorkDir  string      // Working directory override for the build container.
	User     string      // User identity override in "uid:gid" format.
	Security *specs.Spec // OCI runtime spec used to configure the build container's security policy.
	Stdout   io.Writer   // Destination for the command's standard output.
	Stderr   io.Writer   // Destination for the command's standard error.
}

// Image configuration changes applied without adding a new layer.
//
// Only non-zero fields are applied; zero values leave the existing
// configuration unchanged.
type ConfigUpdate struct {
	AddEnv        map[string]string // Environment variables to add or override in the image config.
	SetWorkDir    string            // Working directory to set in the image config.
	SetUser       string            // User identity to set in the image config in "uid:gid" format.
	SetEntrypoint []string          // Entrypoint to set in the image config.
	SetCmd        []string          // Default command to set in the image config.
}
