//go:build darwin || linux

package local

import (
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/crux/internal/runtime"
)

const containerdNamespace = "crucible"

// Returns a [runtime.Runtime] connected to the containerd instance.
func newRuntime(name string) (*runtime.Runtime, error) {
	return runtime.New(paths.ContainerdSocket(name), containerdNamespace)
}
