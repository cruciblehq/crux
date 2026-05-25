package compute

import "github.com/cruciblehq/crux/compute/provider"

// Re-exported so callers can match against [provider.ErrUnknownProvider]
// without taking a direct dependency on the internal provider package.
var ErrUnknownProvider = provider.ErrUnknownProvider
