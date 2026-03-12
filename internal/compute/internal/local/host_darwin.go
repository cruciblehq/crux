//go:build darwin

package local

import (
	"context"
	"path/filepath"
	"runtime"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal"
	"github.com/cruciblehq/crux/internal/compute/internal/provider"
	"github.com/cruciblehq/crux/internal/resource"
	"github.com/cruciblehq/spec/manifest"
)

// Ensures the host VM is running, creating it if necessary.
//
// If the VM does not exist, the machine image is resolved through the given
// [resource.Source], a Lima YAML configuration is generated, and the VM is
// created and started. If the VM exists but is stopped, it is resumed. If
// already running, this is a no-op. Blocks until the VM is ready.
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
		dir, _, err := source.Resolve(ctx, manifest.TypeMachine, internal.DefaultMachineImage)
		if err != nil {
			return crex.Wrap(ErrHostStart, err)
		}

		imagePath := filepath.Join(dir, machineImageForArch())

		configPath, err := generateLimaConfig(name, imagePath)
		if err != nil {
			return err
		}
		if err := limaRun(ctx, "start", "--tty=false", "--name="+limaInstanceName, configPath); err != nil {
			return crex.Wrap(ErrHostStart, err)
		}
		return nil

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

// Returns the machine qcow2 image filename for the host architecture.
func machineImageForArch() string {
	switch runtime.GOARCH {
	case goarchARM64:
		return manifest.MachineImageAarch64
	default:
		return manifest.MachineImageX86_64
	}
}
