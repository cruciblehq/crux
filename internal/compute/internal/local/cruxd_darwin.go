//go:build darwin

package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/paths"
)

// Starts a cruxd process inside the VM for the given instance.
//
// The cruxd process listens on the instance's socket path, which resides
// on the virtiofs mount shared between host and guest. This allows the
// host to connect directly without SSH port forwarding. Any stale socket
// or PID file from a previous run is removed before starting.
func startCruxd(ctx context.Context, name string) error {
	socketPath := paths.CruxdSocket(name)
	pidPath := paths.CruxdPIDFile(name)
	dir := filepath.Dir(socketPath)

	// Remove stale artifacts from a previous run (crash, forced stop).
	os.Remove(socketPath)
	os.Remove(pidPath)

	script := fmt.Sprintf(
		"mkdir -p '%s' && nohup /usr/local/bin/cruxd --socket '%s' --pid-file '%s' >/dev/null 2>&1 &",
		dir, socketPath, pidPath,
	)

	if err := limaShell(ctx, "sudo", "sh", "-c", script); err != nil {
		return crex.Wrap(ErrRuntimeStart, err)
	}
	return nil
}

// Stops the cruxd process for the given instance.
//
// The guest PID is read from the PID file on the virtiofs mount and
// signalled via limactl shell. The socket and PID file are removed
// after the process is stopped.
func stopCruxd(ctx context.Context, name string) error {
	pidPath := paths.CruxdPIDFile(name)

	data, err := os.ReadFile(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrRuntimeNotRunning
		}
		return crex.Wrap(ErrRuntimeStop, err)
	}

	pid := strings.TrimSpace(string(data))
	if err := limaShell(ctx, "sudo", "kill", pid); err != nil {
		return crex.Wrap(ErrRuntimeStop, err)
	}

	os.Remove(paths.CruxdSocket(name))
	os.Remove(pidPath)
	return nil
}
