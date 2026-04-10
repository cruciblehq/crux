//go:build darwin

package local

import (
	"context"
	"errors"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/compute/internal/provider"
	"github.com/cruciblehq/crux/internal/resource"
)

var errMachineProvisioning = errors.New("VM provisioning from derived images is not yet implemented")

// Ensures the host VM is running, creating it if necessary.
//
// If the VM exists but is stopped, it is resumed. If already running, this
// is a no-op. Provisioning a new VM is not yet supported—VM images will be
// derived from blueprint resolution rather than pulled as a standalone
// machine resource.
func ensureHostRunning(ctx context.Context, name string, source resource.Source) error {
	if err := ensureLima(ctx); err != nil {
		return err
	}

	status, err := hostStatus(ctx)
	if err != nil {
		return err
	}

	switch status {
	case provider.StateRunning:
		return nil

	case provider.StateStopped:
		if err := limaRun(ctx, "start", "--tty=false", limaInstanceName); err != nil {
			return crex.Wrap(ErrHostStart, err)
		}
		return nil

	case provider.StateNotProvisioned:
		// TODO: VM images are now derived from blueprint resolution rather
		// than pulled as a standalone machine resource. This path needs to
		// accept a resolved VM image path from the plan layer.
		return crex.Wrap(ErrHostStart, errMachineProvisioning)

	default:
		return crex.Wrapf(ErrHostStart, "unexpected VM state: %s", status)
	}
}

// Deletes the host VM and its disk images.
//
// Blocks until cleanup is complete.
func destroyHost(ctx context.Context) error {
	status, err := hostStatus(ctx)
	if err != nil {
		return err
	}
	if status == provider.StateNotProvisioned {
		return ErrHostNotCreated
	}

	if err := limaRun(ctx, "delete", "--force", limaInstanceName); err != nil {
		return crex.Wrap(ErrHostDestroy, err)
	}
	return nil
}

// Runs a command inside the host VM and captures its output.
func hostExec(ctx context.Context, command string, args ...string) (*provider.ExecResult, error) {
	return limaExec(ctx, command, args...)
}

// Queries the current state of the host VM.
//
// Maps the Lima instance status to a provider state.
func hostStatus(ctx context.Context) (provider.State, error) {
	switch limaInstanceStatus(ctx) {
	case limaStatusRunning:
		return provider.StateRunning, nil
	case limaStatusStopped:
		return provider.StateStopped, nil
	default:
		return provider.StateNotProvisioned, nil
	}
}
