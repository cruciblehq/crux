//go:build linux

package runtime

import (
	"path/filepath"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/cruciblehq/crux/paths"
)

// Returns the containerd socket address.
func containerdAddress() string {
	return filepath.Join(paths.Runtime(), containerdSock)
}

// Creates a containerd client connected to the local containerd socket.
//
// The namespace parameter is used as the containerd namespace.
func newContainerdClient(namespace string) (*containerd.Client, error) {
	return containerd.New(containerdAddress(), containerd.WithDefaultNamespace(namespace))
}
