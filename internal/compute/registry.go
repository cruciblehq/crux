package compute

import (
	"sync"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/compute/internal/local"
	"github.com/cruciblehq/crux/internal/compute/internal/provider"
)

// Identifies a compute backend.
type Provider int

const (
	Local Provider = iota // Host machine (Lima on macOS, native cruxd on Linux).
)

var (
	registryOnce sync.Once        // Ensures the backend registry is only initialised once.
	registry     *backendRegistry // Lazily-initialised singleton backend registry.
)

// Lazily-initialised backend registry.
type backendRegistry struct {
	backends map[Provider]provider.Backend
}

// Returns the lazily-initialised backend registry, creating it if necessary.
func defaultRegistry() *backendRegistry {
	registryOnce.Do(func() {
		registry = &backendRegistry{
			backends: map[Provider]provider.Backend{
				Local: local.NewBackend(),
			},
		}
	})
	return registry
}

// Returns the backend for the given provider.
func BackendFor(p Provider) (Backend, error) {
	r := defaultRegistry()
	b, ok := r.backends[p]
	if !ok {
		return nil, crex.Wrapf(ErrUnknownProvider, "provider %d", p)
	}
	return b, nil
}
