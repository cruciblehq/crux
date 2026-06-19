package compute

import (
	"sync"

	"github.com/cruciblehq/utils-go/crex"
)

var (

	// Ensures the backend registry is only initialised once.
	registryOnce sync.Once

	// Lazily-initialised singleton backend registry.
	registry *backendRegistry
)

// Singleton that maps each [Provider] to its [provider.Backend].
//
// Populated on first call to [defaultRegistry]. Entries are not modified after
// initialisation, so no locking is needed for reads after setup.
type backendRegistry struct {
	backends map[Provider]Backend // Maps each provider to its backend implementation.
}

// Returns the singleton backend registry, initialising it on the first call.
//
// Uses [sync.Once] so concurrent callers block until initialisation completes.
// Subsequent calls return the same registry without acquiring any lock.
func defaultRegistry() *backendRegistry {
	registryOnce.Do(func() {
		registry = &backendRegistry{
			backends: map[Provider]Backend{
				Local: newBackendLocal(),
			},
		}
	})
	return registry
}

// Returns the [Backend] registered for p.
//
// Returns [ErrUnknownProvider] if no backend has been registered for
// the given provider.
func BackendFor(p Provider) (Backend, error) {
	r := defaultRegistry()
	b, ok := r.backends[p]
	if !ok {
		return nil, crex.ProgrammingErrorf("unknown compute provider", "no backend is registered for provider %q", p).
			Cause(ErrUnknownProvider).
			Err()
	}
	return b, nil
}
