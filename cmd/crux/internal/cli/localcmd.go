package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/crux/files"
	"github.com/cruciblehq/crux/manifest"
)

// Name of the lock file used to serialise concurrent blueprint mutations.
const localLockFile = "blueprint.lock"

// Manages the local Crucible environment.
//
// The local environment combines compute infrastructure (start, stop, exec,
// etc.) with the blueprint that tracks which services are registered. Both
// concerns are part of the same environment from the developer's perspective.
type LocalCmd struct {
	Start   LocalStartCmd   `cmd:"" help:"Provision and start the local environment."`
	Stop    LocalStopCmd    `cmd:"" help:"Stop the local environment."`
	Restart LocalRestartCmd `cmd:"" help:"Stop and restart the local environment."`
	Reset   LocalResetCmd   `cmd:"" help:"Destroy and recreate the local environment from scratch."`
	Destroy LocalDestroyCmd `cmd:"" help:"Destroy the local environment and all its data."`
	Status  LocalStatusCmd  `cmd:"" help:"Show local environment status."`
	Exec    LocalExecCmd    `cmd:"" help:"Run a command inside the local environment."`
	Add     LocalAddCmd     `cmd:"" help:"Add a service to the local blueprint."`
	Remove  LocalRemoveCmd  `cmd:"" aliases:"rm" help:"Remove a service from the local blueprint."`
	List    LocalListCmd    `cmd:"" aliases:"ls" help:"List services in the local blueprint."`
	Deploy  LocalDeployCmd  `cmd:"" help:"Build and write the local deployment plan."`
}

// Reads the local blueprint, returning an empty one if none exists yet.
func localBlueprint() (*manifest.Blueprint, error) {
	_, bp, err := openLocalBlueprint()
	return bp, err
}

// Returns compute options from the local deployment plan, if one exists.
//
// Returns a zero-value Options if no plan was written yet or if it cannot be
// decoded. The caller should treat zero values as "no additional requirements"
// and let the backend apply its own defaults.
func localPlanOptions() compute.Options {
	plan, err := manifest.ReadPlanAt(files.BuildDir(files.LocalDir()))
	if err != nil {
		return compute.Options{}
	}
	c, ok := plan.Infrastructure.Computes["default"]
	if !ok {
		return compute.Options{}
	}
	if c.Kernel == nil {
		return compute.Options{}
	}
	return compute.Options{Kernel: *c.Kernel}
}

// Acquires an exclusive lock on the blueprint file, reads it, calls fn, and
// writes the result on success.
//
// All mutations to the local blueprint must go through this function to prevent
// concurrent writes from corrupting the file. The lock is held for the duration
// of fn; callers should keep fn short and avoid blocking I/O inside it.
// No write occurs if fn returns an error.
func modifyLocalBlueprint(ctx context.Context, fn func(*manifest.Blueprint) error) error {
	dir := files.LocalDir()
	if err := os.MkdirAll(dir, files.DefaultDirMode); err != nil {
		return err
	}

	lf, err := os.OpenFile(filepath.Join(dir, localLockFile), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lf.Close()

	if err := files.LockWithContext(ctx, lf); err != nil {
		return err
	}
	defer files.Unlock(lf)

	m, bp, err := openLocalBlueprint()
	if err != nil {
		return err
	}

	if err := fn(bp); err != nil {
		return err
	}

	return manifest.WriteAt(m, dir)
}

// Opens the local blueprint from disk, falling back to an empty in-memory
// blueprint if no blueprint has been written yet.
func openLocalBlueprint() (*manifest.Manifest, *manifest.Blueprint, error) {
	dir := files.LocalDir()
	m, err := manifest.ReadAt(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			bp := &manifest.Blueprint{}
			m := &manifest.Manifest{
				Resource: manifest.Resource{
					Type:    manifest.TypeBlueprint,
					Name:    "default",
					Version: "0.0.0",
				},
				Config: bp,
			}
			return m, bp, nil
		}
		return nil, nil, err
	}
	bp, err := manifest.As[*manifest.Blueprint](m)
	if err != nil {
		return nil, nil, err
	}
	return m, bp, nil
}
