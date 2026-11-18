package compute

import "github.com/cruciblehq/utils-go/crex"

// Provider registered under each name.
var registry = map[ProviderName]Provider{
	Local: localProvider{},
}

// Returns the [Provider] registered under name.
//
// Returns [ErrUnknownProvider] if no provider has that name.
func ProviderFor(name ProviderName) (Provider, error) {
	p, ok := registry[name]
	if !ok {
		return nil, crex.ProgrammingErrorf("unknown compute provider", "no provider is registered as %q", name).
			Cause(ErrUnknownProvider).
			Err()
	}
	return p, nil
}
