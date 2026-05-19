package cgroup

import (
	"strings"

	"github.com/cruciblehq/crux/crex"
)

// Limit for one resource exposed via the cgroup misc controller.
//
// The misc controller is a generic counter for scalar accelerator resources
// (for example sev, sev_es, or vendor-defined keys) that do not warrant a
// dedicated controller of their own.
type misc struct {
	Resource string `knob:"resource" json:"resource,omitempty"` // Resource name.
	Max      uint64 `knob:"max" json:"max,omitempty"`           // Maximum value for the resource.
}

// Parses a misc entry.
//
// Expects a resource name followed by an optional "max=<value>". The
// resource name is required and identifies the entry; an omitted max
// leaves the field at zero, which the caller treats as "not set" when
// reconciling against earlier overrides for the same resource.
func parseMisc(value string) (misc, error) {
	resource, rest, _ := strings.Cut(strings.TrimSpace(value), " ")
	if resource == "" {
		return misc{}, crex.Wrapf(ErrInvalidGrant, "resource name required")
	}
	m := misc{Resource: resource}
	if rest != "" {
		if err := parseArgs(rest, map[string]func(string) error{
			"max": func(v string) error { return parseUint64(&m.Max, v) },
		}); err != nil {
			return misc{}, err
		}
	}
	return m, nil
}

// Whether e and other constrain the same misc resource.
func (e misc) equal(other misc) bool {
	return e.Resource == other.Resource
}

// Returns an error if e and other share identity with conflicting values.
func (e misc) check(other misc) error {
	if !e.equal(other) || e == other {
		return nil
	}
	return crex.Wrapf(ErrConflict, "%s %s already set", "misc", other.Resource)
}

// Leaves e unchanged and always reports no change.
//
// Same-identity entries with conflicting values are rejected upstream by
// check, and identical entries need no merge, so there is nothing to do.
func (e *misc) merge(other misc) bool {
	return false
}
