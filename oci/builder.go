package oci

import (
	"bytes"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/cruciblehq/crex"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// File permission modes for container images.
//
// These constants define the standard permission model for Crucible container
// images. Owner and group can read (and execute where appropriate); others have
// no access. This supports multi-user sandboxed execution where the runtime and
// application run as different users in the same group.
const (
	ModeExecutable os.FileMode = 0550 // Executables and directories: owner+group read/execute
	ModeReadOnly   os.FileMode = 0440 // Regular files: owner+group read only
	modeExecMask   os.FileMode = 0111 // Mask for detecting execute bits
)

// Single-platform OCI image builder.
//
// Accumulates filesystem layers and configuration options for creating an
// OCI-compliant container image. Produces images for a single operating system
// and architecture combination. Uses go-containerregistry under the hood.
type Builder struct {
	image  v1.Image  // Accumulated image with all layers applied.
	config v1.Config // Runtime configuration (entrypoint, env, labels, etc.).
	arch   string    // Target CPU architecture (e.g., "amd64", "arm64").
	os     string    // Target operating system (e.g., "linux", "darwin").
}

// Creates a new image builder for the specified platform.
//
// Initializes a builder configured for the given operating system and CPU
// architecture. The returned builder starts with no layers and default
// configuration values. Callers should add layers and configure runtime options
// before calling SaveTarball to produce the final image.
func NewBuilder(osName, arch string) *Builder {
	return &Builder{
		image: empty.Image,
		arch:  arch,
		os:    osName,
		config: v1.Config{
			Env: []string{},
		},
	}
}

// Creates a builder that extends an existing image.
//
// Initializes a builder starting from the provided base image. The builder
// inherits all layers, configuration, and target platform from the base,
// allowing additional layers and configuration changes to be applied on top.
func NewBuilderFrom(base *Image) (*Builder, error) {
	cfg, err := base.img.ConfigFile()
	if err != nil {
		return nil, crex.Wrap(ErrInvalidImage, err)
	}

	return &Builder{
		image:  base.img,
		arch:   cfg.Architecture,
		os:     cfg.OS,
		config: cfg.Config,
	}, nil
}

// Adds a directory as a new layer.
//
// Recursively walks the source directory and packages all files and directories
// into a single filesystem layer. The destDir parameter specifies the absolute
// path where the contents will appear in the container's filesystem. File
// permissions and directory structure are preserved, but modification times
// are zeroed for reproducible builds.
func (b *Builder) AddDir(srcDir, destDir string) error {
	layerData, err := createTarFromDir(srcDir, destDir)
	if err != nil {
		return crex.Wrap(ErrLayerCreate, err)
	}
	return b.addLayerFromBytes(layerData)
}

// Adds a single file as a new layer.
//
// Reads the source file and creates a layer containing only that file at the
// specified destination path. The mode parameter sets the file's permission
// bits in the container filesystem.
func (b *Builder) AddFile(src, dest string, mode int64) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return crex.Wrap(ErrLayerCreate, err)
	}
	return b.AddBytes(content, dest, mode)
}

// Adds raw bytes as a file in a new layer.
//
// Creates a layer containing a single file with the provided content. Useful
// for embedding generated content, configuration files, or small assets without
// writing them to disk first. The dest parameter specifies the absolute path in
// the container filesystem, and mode sets the permission bits.
func (b *Builder) AddBytes(content []byte, dest string, mode int64) error {
	layerData, err := createTarFromBytes(content, dest, mode)
	if err != nil {
		return crex.Wrap(ErrLayerCreate, err)
	}
	return b.addLayerFromBytes(layerData)
}

// Adds a file or directory with automatic permission detection.
//
// For directories, recursively adds all contents as a layer preserving
// structure. For files, detects whether the source has any execute bit set and
// applies ModeExecutable (0550) or ModeReadOnly (0440) accordingly.
func (b *Builder) AddMapping(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return crex.Wrap(ErrLayerCreate, err)
	}

	if info.IsDir() {
		return b.AddDir(src, dest)
	}

	mode := ModeReadOnly
	if info.Mode()&modeExecMask != 0 {
		mode = ModeExecutable
	}
	return b.AddFile(src, dest, int64(mode))
}

// Appends tar archive data as a new layer to the image.
func (b *Builder) addLayerFromBytes(data []byte) error {
	opener := func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	layer, err := tarball.LayerFromOpener(opener)
	if err != nil {
		return crex.Wrap(ErrLayerCreate, err)
	}
	b.image, err = mutate.AppendLayers(b.image, layer)
	if err != nil {
		return crex.Wrap(ErrLayerCreate, err)
	}
	return nil
}

// Sets the image entrypoint.
//
// Configures the executable that runs when the container starts. The entrypoint
// is specified as a list of strings where the first element is the command and
// subsequent elements are fixed arguments. Setting an entrypoint clears any
// previously configured default command.
func (b *Builder) SetEntrypoint(entrypoint ...string) {
	b.config.Entrypoint = entrypoint
	b.config.Cmd = nil
}

// Sets the default command arguments.
//
// Configures the default arguments passed to the entrypoint when no command is
// specified at container runtime. If no entrypoint is set, the first element of
// cmd becomes the executable.
func (b *Builder) SetCmd(cmd ...string) {
	b.config.Cmd = cmd
}

// Adds an environment variable to the image configuration.
//
// Appends a new environment variable that will be set in the container's
// execution environment. The variable is stored in KEY=value format. Multiple
// calls accumulate variables; there is no way to remove a previously added
// variable through this interface.
func (b *Builder) SetEnv(key, value string) {
	b.config.Env = append(b.config.Env, key+"="+value)
}

// Sets the working directory for the container process.
//
// Configures the initial current directory when the container starts. This
// affects where relative paths resolve and where the entrypoint process begins
// execution. The directory must exist in the container's filesystem, typically
// created by one of the added layers.
func (b *Builder) SetWorkdir(dir string) {
	b.config.WorkingDir = dir
}

// Sets a metadata label on the image.
//
// Adds a key-value pair to the image's label map. Labels are arbitrary metadata
// that can be used for organization, filtering, or conveying build information.
// Common conventions include version numbers, maintainer contacts, and source
// repository URLs.
func (b *Builder) SetLabel(key, value string) {
	if b.config.Labels == nil {
		b.config.Labels = make(map[string]string)
	}
	b.config.Labels[key] = value
}

// Applies runtime configuration and returns the finalized image.
func (b *Builder) build() (v1.Image, error) {
	cfg, err := b.image.ConfigFile()
	if err != nil {
		return nil, crex.Wrap(ErrImageBuild, err)
	}

	cfg.Config = b.config
	cfg.Architecture = b.arch
	cfg.OS = b.os

	img, err := mutate.ConfigFile(b.image, cfg)
	if err != nil {
		return nil, crex.Wrap(ErrImageBuild, err)
	}
	return img, nil
}

// Writes the image to an OCI tarball file.
//
// Serializes the accumulated layers and configuration into a complete OCI image
// layout and writes it as a tar archive to the specified path. The resulting
// file can be loaded into container runtimes or pushed to OCI-compliant
// registries. This method creates or overwrites the destination file.
func (b *Builder) SaveTarball(path string) error {
	img, err := b.build()
	if err != nil {
		return err
	}

	idx := mutate.AppendManifests(empty.Index, mutate.IndexAddendum{
		Add: img,
		Descriptor: v1.Descriptor{
			Platform: &v1.Platform{
				Architecture: b.arch,
				OS:           b.os,
			},
		},
	})

	return writeIndexToTarball(idx, path)
}

// Returns the built image with runtime configuration applied.
//
// Finalizes the accumulated layers and configuration into an [Image] that can
// be passed to [NewBuilderFrom] for extension, added to a
// [MultiPlatformBuilder] via [MultiPlatformBuilder.AddImage], or inspected
// for its digest and layers.
func (b *Builder) Image() (*Image, error) {
	img, err := b.build()
	if err != nil {
		return nil, err
	}
	return &Image{img: img}, nil
}

// Multi-architecture OCI image builder.
//
// Manages multiple single-platform builders to create images that support
// different operating system and architecture combinations. Each platform has
// its own set of layers and configuration, but they share a common image index.
// This enables container runtimes to automatically select the appropriate
// platform-specific image based on the host system.
type MultiPlatformBuilder struct {
	builders map[string]*Builder // Platform-specific builders keyed by "os/arch" (e.g., "linux/amd64").
	images   map[string]*Image   // Pre-built images keyed by "os/arch".
}

// Creates a builder for multi-platform images.
//
// Initializes an empty multi-platform builder with no configured platforms.
// Callers must use ForPlatform to add platform-specific builders before calling
// SaveTarball. The builder can support any number of platform combinations.
func NewMultiPlatformBuilder() *MultiPlatformBuilder {
	return &MultiPlatformBuilder{
		builders: make(map[string]*Builder),
		images:   make(map[string]*Image),
	}
}

// Returns a builder for the specified platform.
//
// Retrieves or creates a single-platform builder for the given operating system
// and architecture. Subsequent calls with the same platform return the same
// builder instance, allowing incremental configuration. The returned builder
// can be used to add layers and set configuration specific to that platform.
func (mb *MultiPlatformBuilder) ForPlatform(osName, arch string) *Builder {
	key := osName + "/" + arch
	if b, ok := mb.builders[key]; ok {
		return b
	}
	b := NewBuilder(osName, arch)
	mb.builders[key] = b
	return b
}

// Adds a pre-built image for the specified platform.
//
// Use this when extending an existing base image rather than building from
// scratch. The image will be included in the final multi-platform index
// alongside any images created via ForPlatform.
func (mb *MultiPlatformBuilder) AddImage(osName, arch string, img *Image) {
	key := osName + "/" + arch
	mb.images[key] = img
}

// Writes a multi-platform OCI image to a tarball.
//
// Serializes all platform-specific images into a single OCI image layout with
// an index referencing each platform's manifest. Layers that are identical
// across platforms are deduplicated in the output to reduce file size. The
// resulting tarball can be pushed to registries that support multi-architecture
// images or loaded into container runtimes that select the appropriate platform
// automatically.
func (mb *MultiPlatformBuilder) SaveTarball(path string) error {
	idx, err := mb.buildIndex()
	if err != nil {
		return err
	}
	return writeIndexToTarball(idx, path)
}

// Returns the multi-platform index for registry operations.
//
// Assembles all platform-specific images into an [Index] containing every
// configured platform. The returned index can be inspected for its digest or
// saved to an OCI layout directory via [Index.SaveLayout]. Platforms are
// ordered deterministically by their "os/arch" key.
func (mb *MultiPlatformBuilder) Index() (*Index, error) {
	idx, err := mb.buildIndex()
	if err != nil {
		return nil, err
	}
	return &Index{idx: idx}, nil
}

// Assembles all platform images into an index.
func (mb *MultiPlatformBuilder) buildIndex() (v1.ImageIndex, error) {
	var idx v1.ImageIndex = empty.Index

	// Iterate in sorted key order for deterministic output.
	for _, key := range sortedKeys(mb.builders) {
		b := mb.builders[key]
		img, err := b.Image()
		if err != nil {
			return nil, err
		}

		idx = mutate.AppendManifests(idx, mutate.IndexAddendum{
			Add: img.img,
			Descriptor: v1.Descriptor{
				Platform: &v1.Platform{
					Architecture: b.arch,
					OS:           b.os,
				},
			},
		})
	}

	for _, key := range sortedKeys(mb.images) {
		img := mb.images[key]
		parts := strings.Split(key, "/")
		osName, arch := parts[0], parts[1]

		idx = mutate.AppendManifests(idx, mutate.IndexAddendum{
			Add: img.img,
			Descriptor: v1.Descriptor{
				Platform: &v1.Platform{
					Architecture: arch,
					OS:           osName,
				},
			},
		})
	}

	return idx, nil
}

// Returns the keys of a map in sorted order.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
