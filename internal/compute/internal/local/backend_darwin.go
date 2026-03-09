//go:build darwin

package local

import (
	"context"
	"log/slog"
	"net"

	"github.com/cruciblehq/crux/internal/compute/internal/provider"
	"github.com/cruciblehq/crux/internal/paths"
)

// Provisions a cruxd instance.
//
// The shared VM is created and started if it does not already exist. A cruxd
// process is then started inside the VM for the given instance. The call
// blocks until cruxd signals readiness via the ready-fd protocol.
func provision(ctx context.Context, config *provider.Config) error {
	if err := ensureRuntimeRunning(ctx, config.Version); err != nil {
		slog.Debug("provision failed, VM did not start", "name", config.Name, "error", err)
		return err
	}

	if err := startCruxd(ctx, config.Name); err != nil {
		slog.Debug("provision failed, cruxd did not start", "name", config.Name, "error", err)
		return err
	}

	return nil
}

// Starts a cruxd instance inside an already-running VM.
//
// The VM must have been provisioned and be running. A cruxd process is
// started inside the VM. The call blocks until cruxd signals readiness
// via the ready-fd protocol.
func start(ctx context.Context, name string) error {
	state, err := runtimeStatus(ctx)
	if err != nil {
		return err
	}
	if state != provider.StateRunning {
		slog.Debug("start failed, VM is not running", "name", name, "state", state)
		return ErrRuntimeNotRunning
	}

	if err := startCruxd(ctx, name); err != nil {
		slog.Debug("start failed, cruxd did not start", "name", name, "error", err)
		return err
	}

	return nil
}

// Stops a cruxd instance. The VM continues running.
func stop(ctx context.Context, name string) error {
	return stopCruxd(ctx, name)
}

// Tears down a cruxd instance and destroys the VM.
func deprovision(ctx context.Context, name string) error {
	// Best-effort: stop the cruxd instance if running.
	stopCruxd(ctx, name)

	return destroyRuntime(ctx)
}

// Queries the current state of a cruxd instance.
//
// If the runtime has not been provisioned, the instance is in a
// [provider.StateNotProvisioned] state. If the runtime exists but is not
// running (or the cruxd process is not reachable), the instance is in a
// [provider.StateStopped] state. If the cruxd socket is reachable, the
// instance is in a [provider.StateRunning] state.
func status(ctx context.Context, name string) (provider.State, error) {
	rtState, err := runtimeStatus(ctx)
	if err != nil {
		return 0, err
	}

	if rtState == provider.StateNotProvisioned {
		return provider.StateNotProvisioned, nil
	}
	if rtState != provider.StateRunning {
		return provider.StateStopped, nil
	}

	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", paths.CruxdSocket(name))
	if err != nil {
		return provider.StateStopped, nil
	}
	conn.Close()
	return provider.StateRunning, nil
}

// Runs a command inside the runtime VM.
func execute(ctx context.Context, _ string, command string, args ...string) (*provider.ExecResult, error) {
	return runtimeExec(ctx, command, args...)
}
