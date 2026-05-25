//go:build darwin || linux

package local

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"github.com/cruciblehq/crux/compute/provider"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/paths"
)

// Provisions a compute instance.
//
// The shared VM is created and started if it does not already exist. The
// containerd runs as a system service inside the VM and is started
// automatically during boot.
func provision(ctx context.Context, name string) error {
	return ensureHostRunning(ctx, name)
}

// Starts the VM for a previously provisioned instance.
//
// The VM must have been provisioned already. containerd starts automatically
// when the VM boots.
func start(ctx context.Context, _ string) error {
	state, err := hostStatus(ctx)
	if err != nil {
		return err
	}
	if state == provider.StateRunning {
		return nil
	}
	if state == provider.StateNotProvisioned {
		return ErrHostNotCreated
	}

	if err := limaRun(ctx, "start", "--tty=false", limaInstanceName); err != nil {
		return crex.Wrap(ErrHostStart, err)
	}
	return nil
}

// Stops the VM.
func stop(ctx context.Context, _ string) error {
	state, err := hostStatus(ctx)
	if err != nil {
		return err
	}
	if state != provider.StateRunning {
		return ErrHostNotRunning
	}

	if err := limaRun(ctx, "stop", limaInstanceName); err != nil {
		return crex.Wrap(ErrHostStop, err)
	}
	return nil
}

// Tears down the instance and destroys the VM.
//
// Returns nil if the VM does not exist (idempotent).
func deprovision(ctx context.Context, _ string) error {
	err := destroyHost(ctx)
	if errors.Is(err, ErrHostNotCreated) {
		return nil
	}
	return err
}

// Queries the current state of a compute instance.
//
// State is determined by probing two layers: the Lima VM and the containerd
// socket inside it. The returned state is the least-healthy of the two:
//
//   - [provider.StateNotProvisioned] — the VM does not exist.
//   - [provider.StateStopped] — the VM exists but is not running, or the
//     VM is running but the containerd socket is not reachable.
//   - [provider.StateRunning] — both the VM and containerd are up.
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
	conn, err := d.DialContext(ctx, "unix", paths.ContainerdSocket(limaInstanceName))
	if err != nil {
		return provider.StateStopped, nil
	}

	// Lima's port forwarding accepts connections to the host socket even
	// when the guest socket has no listener, returning EOF immediately.
	// Probe the connection to distinguish a live containerd from a
	// forwarding stub: containerd will send gRPC data (timeout), while a
	// dead-end connection returns EOF.
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, err = conn.Read(make([]byte, 1))
	conn.Close()
	if err == io.EOF {
		return provider.StateStopped, nil
	}
	return provider.StateRunning, nil
}

// Runs a command inside the host VM.
func execute(ctx context.Context, _ string, command string, args ...string) (*provider.ExecResult, error) {
	return hostExec(ctx, command, args...)
}

// Blocks until containerd inside the VM is accepting connections.
//
// Polls the forwarded containerd socket every two seconds. Returns an error
// if the context is cancelled or the three-minute deadline is exceeded.
// Lima's port-forward proxy accepts connections before the SSH tunnel is
// fully established and responds with EOF; a timeout on the read indicates
// containerd is alive and waiting for gRPC.
func waitForContainerd(ctx context.Context) error {
	deadline := time.Now().Add(15 * time.Minute)
	for {
		if isContainerdReady(ctx) {
			return nil
		}
		if time.Now().After(deadline) {
			return crex.Wrapf(ErrHostStart, "timed out waiting for containerd to become ready")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

// Reports whether the forwarded containerd socket is accepting real connections.
func isContainerdReady(ctx context.Context) bool {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", paths.ContainerdSocket(limaInstanceName))
	if err != nil {
		return false
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, err = conn.Read(make([]byte, 1))
	// EOF means the Lima proxy accepted the connection but the SSH tunnel
	// is not yet established. A timeout means containerd is waiting for
	// gRPC data, which is the live state we want.
	return err != io.EOF
}
