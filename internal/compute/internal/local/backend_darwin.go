//go:build darwin

package local

import (
	"context"
	"log/slog"
	"net"

	"github.com/cruciblehq/crux/internal/compute/internal/provider"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/resource"
)

// Provisions a cruxd instance.
//
// The shared VM is created and started if it does not already exist. The
// machine image is resolved through the [resource.Source] passed by the
// caller. A cruxd process is then started inside the VM for the given
// instance. The call blocks until cruxd signals readiness via the
// ready-fd protocol.
func provision(ctx context.Context, name string, source resource.Source) error {
	if err := ensureHostRunning(ctx, name, source); err != nil {
		return err
	}

	if err := startCruxd(ctx, name); err != nil {
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
	state, err := hostStatus(ctx)
	if err != nil {
		return err
	}
	if state != provider.StateRunning {
		slog.Debug("start failed, VM is not running", "name", name, "state", state)
		return ErrHostNotRunning
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

	return destroyHost(ctx)
}

// Queries the current state of a cruxd instance.
//
// State is determined by probing two layers: the Lima VM and the cruxd
// process inside it. The returned state is the least-healthy of the two:
//
//   - [provider.StateNotProvisioned] — the VM does not exist.
//   - [provider.StateStopped] — the VM exists but is not running, or the
//     VM is running but the cruxd socket for this instance is not reachable.
//   - [provider.StateRunning] — both the VM and the cruxd socket are up.
func status(ctx context.Context, name string) (provider.State, error) {
	rtState, err := hostStatus(ctx)
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

// Runs a command inside the host VM.
func execute(ctx context.Context, _ string, command string, args ...string) (*provider.ExecResult, error) {
	return hostExec(ctx, command, args...)
}
