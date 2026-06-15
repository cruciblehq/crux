package fcap

import "github.com/cruciblehq/crux/crex"

// Accumulated file capability spec, keyed by binary path.
//
// Multiple affordances targeting the same binary path are merged into a single
// entry so the resulting map holds the minimal set of grants needed by the
// workload.
type Spec struct {

	// Per-binary file capabilities.
	//
	// The key is the absolute path to the binary. Each value is the accumulated
	// capabilities to grant on that binary.
	Entries map[string]*Capabilities `codec:"entries"`
}

// Validates the fcap spec.
//
// Each path key must be non-empty and its capabilities must be valid.
func (s *Spec) Validate() error {
	for path, caps := range s.Entries {
		if path == "" {
			return crex.Wrapf(ErrInvalidFcap, "empty path key")
		}
		if caps == nil {
			return crex.Wrapf(ErrInvalidFcap, "nil capabilities for path %q", path)
		}
		if err := caps.Validate(); err != nil {
			return crex.Wrap(ErrInvalidFcap, err)
		}
	}
	return nil
}
