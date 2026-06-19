package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/cruciblehq/crux/compute"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Name of the lock file used to serialise concurrent blueprint mutations.
const localLockFile = "blueprint.lock"

// Path to the local blueprint state directory.
func localStateDir() string {
	return filepath.Join(xdg.DataHome, "crux", "local")
}

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
	plan, err := manifest.ReadPlanAt(file.BuildDir(localStateDir()))
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
	const description = "cannot access local state"
	const recoveryWriteAccess = "Make sure you have write access to %s, then try again."

	dir := localStateDir()
	if err := os.MkdirAll(dir, file.DefaultDirMode); err != nil {
		return crex.SystemError(description, "failed to create the local state directory").
			Recoveryf(recoveryWriteAccess, dir).
			Cause(err).
			Err()
	}

	lf, err := os.OpenFile(filepath.Join(dir, localLockFile), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return crex.SystemError(description, "failed to open the local state lock").
			Recoveryf(recoveryWriteAccess, dir).
			Cause(err).
			Err()
	}
	defer lf.Close()

	if err := file.LockWithContext(ctx, lf); err != nil {
		return crex.SystemError(description, "failed to lock the local state").
			Recoveryf("Another crux process may be holding the lock at %s; wait for it to finish and try again.", filepath.Join(dir, localLockFile)).
			Cause(err).
			Err()
	}
	defer file.Unlock(lf)

	m, bp, err := openLocalBlueprint()
	if err != nil {
		return err
	}

	if err := fn(bp); err != nil {
		return err
	}

	if err := manifest.WriteAt(m, dir); err != nil {
		return crex.SystemError("cannot update local state", "failed to write the local blueprint").
			Recoveryf("Make sure you have write access to %s, then try again.", dir).
			Cause(err).
			Err()
	}
	return nil
}

// Opens the local blueprint from disk, falling back to an empty in-memory
// blueprint if no blueprint has been written yet.
func openLocalBlueprint() (*manifest.Manifest, *manifest.Blueprint, error) {
	dir := localStateDir()
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
		return nil, nil, crex.SystemError("cannot read local state", "the local blueprint could not be read").
			Recoveryf("Make sure you have read access to %s, then try again.", dir).
			Cause(err).
			Err()
	}
	bp, err := manifest.As[*manifest.Blueprint](m)
	if err != nil {
		return nil, nil, crex.SystemError("cannot read local state", "the local blueprint is malformed").
			Recoveryf("The local state at %s is corrupt and may need to be removed manually.", dir).
			Cause(err).
			Err()
	}
	return m, bp, nil
}
