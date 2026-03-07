package compute

import "github.com/cruciblehq/crux/internal/compute/internal/provider"

var (
	ErrUnknownProvider = provider.ErrUnknownProvider

	ErrConnectionFailed  = provider.ErrConnectionFailed
	ErrConnectionRefused = provider.ErrConnectionRefused
	ErrRequestFailed     = provider.ErrRequestFailed
	ErrNotRunning        = provider.ErrNotRunning
)
