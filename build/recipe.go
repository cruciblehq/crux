package build

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/cruciblehq/crux/cache"
	"github.com/cruciblehq/crux/kit/archive"
	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/pack"
	"github.com/cruciblehq/crux/pull"
	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
	"github.com/cruciblehq/crux/runtime"
)

// Holds shared state for building all stages of a recipe.
type recipeBuild struct {
	client  *containerd.Client          // Shared containerd gRPC connection.
	id      *reference.Identifier       // Parsed resource identifier.
	version string                      // Resource version from the manifest, used as the base for stage image versions.
	options reference.IdentifierOptions // Options for parsing references in the recipe.
	output  string                      // Output directory for the final build artifact.
	context string                      // Directory containing the manifest, root for resolving copy sources.
}

// Builds a recipe end-to-end against the container runtime.
//
// This is the shared pipeline for all resource types that embed a recipe.
// All stages are built in declaration order. The non-transient stage is
// exported as the final image.
func buildRecipe(ctx context.Context, m manifest.Manifest, recipe *manifest.Recipe, registry, defaultNamespace, output, context string) (*Result, error) {
	options := reference.IdentifierOptions{
		DefaultRegistry:  registry,
		DefaultNamespace: defaultNamespace,
	}

	id, err := reference.ParseIdentifier(m.Resource.Name, m.Resource.Type, options)
	if err != nil {
		return nil, err
	}

	if err := ensureRuntime(); err != nil {
		return nil, err
	}

	client, err := runtime.NewContainerdClient(id.Hostname())
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ctx, done, err := client.WithLease(ctx)
	if err != nil {
		return nil, err
	}
	defer done(ctx)

	rb := &recipeBuild{
		client:  client,
		id:      id,
		version: m.Resource.Version,
		options: options,
		output:  output,
		context: context,
	}

	for i, stage := range recipe.Stages {
		if err := rb.buildStage(ctx, stage, i); err != nil {
			return nil, fmt.Errorf("stage %s: %w", stageLabel(stage.Name, i), err)
		}
	}

	return &Result{Output: output, Manifest: &m}, nil
}

// Builds a single stage of a recipe.
//
// Resolves the stage's base image, starts a build container, executes the
// stage's steps, then commits the result. Non-transient stages are exported
// to the output directory.
func (rb *recipeBuild) buildStage(ctx context.Context, stage manifest.Stage, index int) error {
	label := stageLabel(stage.Name, index)
	slog.Info("building stage", "stage", label)

	source, err := parseFrom(stage.From, rb.options)
	if err != nil {
		return err
	}

	version := rb.version
	if !stage.Transient {
		version = stageVersion(version, label)
	}
	img := runtime.NewImage(rb.client, rb.id, version)

	if err := rb.resolveSource(ctx, img, source); err != nil {
		return err
	}

	ctr, err := img.Start(ctx, "")
	if err != nil {
		return err
	}
	defer ctr.Destroy(ctx)

	if err := executeSteps(ctx, ctr, stage.Steps, newRecipeState(), rb.context); err != nil {
		return err
	}

	if err := ctr.Stop(ctx); err != nil {
		return err
	}

	if err := ctr.Commit(ctx); err != nil {
		return err
	}

	if !stage.Transient {
		return img.Export(ctx, filepath.Join(rb.output, pack.ImageFile))
	}

	return nil
}

// Resolves a recipe source into an imported container image.
//
// For file sources the local OCI tarball is imported directly. For ref
// sources the runtime archive is pulled from the registry (with caching),
// extracted to a temporary directory, and the contained image.tar is
// imported.
func (rb *recipeBuild) resolveSource(ctx context.Context, img *runtime.Image, src source) error {
	switch src.Type {
	case sourceFile:
		return img.Import(ctx, src.Value)
	case sourceRef:
		return rb.resolveRefSource(ctx, img, src)
	}
	return nil
}

// Pulls a runtime archive from the registry, extracts image.tar, and imports
// it into the container runtime.
func (rb *recipeBuild) resolveRefSource(ctx context.Context, img *runtime.Image, src source) error {
	result, err := pull.Pull(ctx, pull.Options{
		Registry:         rb.options.DefaultRegistry,
		Reference:        src.Value,
		Type:             resource.TypeRuntime,
		DefaultNamespace: rb.options.DefaultNamespace,
	})
	if err != nil {
		return err
	}

	localCache, err := cache.Open(ctx, nil)
	if err != nil {
		return err
	}
	defer localCache.Close()

	archiveReader, err := localCache.OpenArchive(ctx, result.Namespace, result.Resource, result.Version)
	if err != nil {
		return err
	}
	defer archiveReader.Close()

	extractDir, err := os.MkdirTemp("", "crux-runtime-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(extractDir)

	if err := archive.ExtractFromReader(archiveReader, extractDir, archive.Zstd); err != nil {
		return err
	}

	imagePath := filepath.Join(extractDir, pack.ImageFile)
	return img.Import(ctx, imagePath)
}

// Returns a version string for a transient stage image, distinguishing it
// from the output image.
func stageVersion(version, label string) string {
	return fmt.Sprintf("%s-stage-%s", version, label)
}

// Returns a human-readable label for a stage, preferring the name when
// available and falling back to the 1-based index.
func stageLabel(name string, index int) string {
	if name != "" {
		return fmt.Sprintf("%q", name)
	}
	return fmt.Sprintf("%d", index+1)
}

// Executes a list of steps in order against the build container.
func executeSteps(ctx context.Context, ctr *runtime.Container, steps []manifest.Step, state *recipeState, context string) error {
	for i, step := range steps {
		if err := executeStep(ctx, ctr, step, state, context); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}
	return nil
}

// Executes a single step, dispatching to operation execution, group recursion,
// or state mutation depending on the step's fields.
func executeStep(ctx context.Context, ctr *runtime.Container, step manifest.Step, state *recipeState, context string) error {
	hasOp := step.Run != "" || step.Copy != ""

	// Platform group: apply group-level modifiers and recurse.
	if len(step.Steps) > 0 {
		state.apply(step)
		return executeSteps(ctx, ctr, step.Steps, state, context)
	}

	// Operation with optional scoped modifiers.
	if hasOp {
		return executeOperation(ctx, ctr, step, state, context)
	}

	// Standalone modifier(s): persist in state.
	state.apply(step)
	return nil
}

// Executes a run or copy operation with scoped modifier overrides.
//
// Step-level modifiers override the persistent state for this operation only.
// The persistent state is not modified.
func executeOperation(ctx context.Context, ctr *runtime.Container, step manifest.Step, state *recipeState, context string) error {
	shell, opts := state.resolve(step)

	if err := ensureDir(ctx, ctr, opts.Workdir); err != nil {
		return err
	}

	switch {
	case step.Run != "":
		slog.Debug("run", "command", step.Run, "shell", shell)
		result, err := ctr.ExecWith(ctx, opts, shell, "-c", step.Run)
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("command failed (exit %d): %s", result.ExitCode, result.Stderr)
		}

	case step.Copy != "":
		if err := executeCopy(ctx, ctr, step.Copy, opts.Workdir, context); err != nil {
			return err
		}
	}

	return nil
}

// Executes a copy operation, transferring host files into the container.
//
// The copy string has the format "src dest", where src is a path on the host
// resolved relative to context and dest is an absolute path inside the
// container. The source can be a file or directory; directories are copied
// recursively. The source is streamed into the container as a tar archive
// via [runtime.Container.CopyTo].
func executeCopy(ctx context.Context, ctr *runtime.Container, copyStr, workdir, context string) error {
	src, dest, err := parseCopy(copyStr, workdir)
	if err != nil {
		return crex.Wrap(ErrCopy, err)
	}

	// Resolve source relative to the manifest directory.
	if !filepath.IsAbs(src) {
		src = filepath.Join(context, src)
	}

	info, err := os.Stat(src)
	if err != nil {
		return crex.Wrap(ErrCopy, err)
	}

	slog.Debug("copy", "src", src, "dest", dest, "dir", info.IsDir())

	// Ensure the destination parent directory exists.
	if err := ensureDir(ctx, ctr, filepath.Dir(dest)); err != nil {
		return crex.Wrap(ErrCopy, err)
	}

	pr, pw := io.Pipe()

	go func() {
		tw := tar.NewWriter(pw)
		var writeErr error

		if info.IsDir() {
			writeErr = writeDirToTar(tw, src, filepath.Base(dest))
		} else {
			writeErr = writeFileToTar(tw, src, filepath.Base(dest))
		}

		tw.Close()
		pw.CloseWithError(writeErr)
	}()

	if err := ctr.CopyTo(ctx, pr, filepath.Dir(dest)); err != nil {
		return crex.Wrap(ErrCopy, err)
	}

	return nil
}

// Parses a copy string into source and destination paths.
//
// The string must contain exactly two whitespace-separated tokens. If dest
// is not absolute, it is joined with workdir.
func parseCopy(s, workdir string) (src, dest string, err error) {
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected source and destination, got %q", s)
	}

	src = parts[0]
	dest = parts[1]

	if !filepath.IsAbs(dest) {
		if workdir == "" {
			return "", "", fmt.Errorf("relative dest %q requires workdir", dest)
		}
		dest = filepath.Join(workdir, dest)
	}

	return src, dest, nil
}

// Writes a single file to a tar writer with the given archive name.
func writeFileToTar(tw *tar.Writer, hostPath, name string) error {
	info, err := os.Stat(hostPath)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = name
	header.Mode = int64(archive.FileMode)

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	f, err := os.Open(hostPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(tw, f)
	return err
}

// Writes a directory tree to a tar writer rooted at the given archive prefix.
func writeDirToTar(tw *tar.Writer, hostDir, prefix string) error {
	return filepath.WalkDir(hostDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(hostDir, path)
		if err != nil {
			return err
		}

		archivePath := filepath.ToSlash(filepath.Join(prefix, relPath))
		return writeTarEntry(tw, path, archivePath, d)
	})
}

// Writes a single file or directory entry to a tar writer.
func writeTarEntry(tw *tar.Writer, hostPath, archivePath string, d os.DirEntry) error {
	info, err := d.Info()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = archivePath

	if info.IsDir() {
		header.Mode = int64(archive.DirMode)
	} else {
		header.Mode = int64(archive.FileMode)
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if info.Mode().IsRegular() {
		f, err := os.Open(hostPath)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	}

	return nil
}

// Creates a directory inside the container. No-op if dir is empty.
func ensureDir(ctx context.Context, ctr *runtime.Container, dir string) error {
	if dir == "" {
		return nil
	}
	result, err := ctr.Exec(ctx, "mkdir", "-p", dir)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to create workdir %q (exit %d): %s", dir, result.ExitCode, result.Stderr)
	}
	return nil
}

// Starts the container runtime if it is not already running.
func ensureRuntime() error {
	status, err := runtime.Status()
	if err != nil {
		return err
	}
	if status == runtime.StateRunning {
		return nil
	}
	slog.Info("starting runtime")
	return runtime.Start()
}
