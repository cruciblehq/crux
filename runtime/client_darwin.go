//go:build darwin

package runtime

import (
	"path/filepath"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/cruciblehq/crux/paths"
)

// Returns the path to the containerd socket.
//
// Lima's portForwards configuration tunnels the guest socket at
// /run/containerd/containerd.sock to this host path.
func containerdForwardedSocket() string {
	return filepath.Join(paths.VM(), containerdSock)
}

// Creates a containerd client connected to the Lima-forwarded socket.
//
// The namespace parameter is used as the containerd namespace.
func newContainerdClient(namespace string) (*containerd.Client, error) {
	return containerd.New(containerdForwardedSocket(), containerd.WithDefaultNamespace(namespace))
}
