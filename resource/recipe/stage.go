package recipe

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/resource/oci"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Persistent mutable build state within a single stage.
//
// Tracks the cumulative effect of standalone modifier steps (shell, workdir,
// user, env) that persist across subsequent steps in the same stage.
type stageState struct {
	shell   string            // Current shell executable, empty means /bin/sh.
	workdir string            // Current working directory override, empty uses image default.
	user    string            // Current user override in uid:gid format, empty uses image default.
	env     map[string]string // Accumulated environment overrides for subsequent steps.
}

// Runs a single stage.
//
// Imports the base image and executes each step in order. stageImages is
// updated with the final committed image ref under the stage name when the
// stage has a name, making it available as a copy source for later stages.
// The caller is responsible for closing the returned container.
func (b *Builder) runStage(ctx context.Context, num int, stage *manifest.Stage, stageImages map[string]string) (*compute.Container, error) {
	ctr, err := b.importBase(ctx, stage, compute.RuntimeOptions{})
	if err != nil {
		return nil, crex.At(crex.Wrap(ErrBuild, err), "stage", num)
	}

	state := &stageState{}
	for j := range stage.Steps {
		if err := b.executeStep(ctx, ctr, &stage.Steps[j], state, stageImages); err != nil {
			ctr.Destroy(ctx)
			return nil, crex.At(crex.At(crex.Wrap(ErrBuild, err), "step", j+1), "stage", num)
		}
	}

	if stage.Name != "" {
		img, err := ctr.Commit(ctx, stage.Name)
		if err != nil {
			ctr.Destroy(ctx)
			return nil, crex.At(crex.Wrap(ErrBuild, err), "stage", num)
		}
		stageImages[stage.Name] = img
	}

	return ctr, nil
}

// Resolves and imports the base image for a stage.
//
// When stage.From is nil the stage starts from scratch and a minimal empty
// OCI image is imported. When From is set, the referenced runtime is pulled
// and its image.tar is imported into the compute backend.
func (b *Builder) importBase(ctx context.Context, stage *manifest.Stage, opts compute.RuntimeOptions) (*compute.Container, error) {
	if stage.From == "" {
		return b.importScratch(ctx, opts)
	}

	ref, err := b.src.Parse(string(manifest.TypeRuntime), stage.From)
	if err != nil {
		return nil, err
	}

	result, err := b.src.Pull(ctx, ref)
	if err != nil {
		return nil, err
	}

	imgPath := filepath.Join(result.Extracted, file.ImageFile)
	f, err := os.Open(imgPath)
	if err != nil {
		return nil, crex.Wrap(oci.ErrFileSystemOperation, err)
	}
	defer f.Close()

	imgRef, err := b.client.Import(ctx, f)
	if err != nil {
		return nil, err
	}
	return b.client.Load(ctx, imgRef, opts)
}

// Creates and imports a minimal empty (scratch) OCI image.
//
// The image has no layers and an empty configuration. It is imported into the
// compute backend and a container is opened from it.
func (b *Builder) importScratch(ctx context.Context, opts compute.RuntimeOptions) (*compute.Container, error) {
	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(oci.WriteScratchTar(pw))
	}()
	ref, err := b.client.Import(ctx, pr)
	if err != nil {
		return nil, err
	}
	return b.client.Load(ctx, ref, opts)
}

