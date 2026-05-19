package cgroup

import "github.com/cruciblehq/crux/crex"

// Per-device I/O bandwidth and IOPS caps.
//
// Imposes hard read/write throughput limits on one block device,
// independent of any weight-based proportional sharing.
type ioMax struct {
	Major uint32 `knob:"major" json:"major,omitempty"` // Device major number.
	Minor uint32 `knob:"minor" json:"minor,omitempty"` // Device minor number.
	Rbps  uint64 `knob:"rbps" json:"rbps,omitempty"`   // Read bandwidth limit in bytes per second.
	Wbps  uint64 `knob:"wbps" json:"wbps,omitempty"`   // Write bandwidth limit in bytes per second.
	Riops uint64 `knob:"riops" json:"riops,omitempty"` // Read IOPS limit in IOPS per second.
	Wiops uint64 `knob:"wiops" json:"wiops,omitempty"` // Write IOPS limit in IOPS per second.
}

// Parses an io.max entry.
//
// Expects "major:minor" followed by zero or more space-separated key=value
// pairs drawn from rbps, wbps, riops, and wiops. The device identity is
// required; omitted caps remain at zero and are treated as "not set" when
// reconciling against earlier overrides for the same device.
func parseIOMax(value string) (ioMax, error) {
	var m ioMax
	var rest string
	var err error
	m.Major, m.Minor, rest, err = parseMajorMinor(value)
	if err != nil {
		return ioMax{}, err
	}
	if rest != "" {
		if err := parseArgs(rest, map[string]func(string) error{
			"rbps":  func(v string) error { return parseUint64(&m.Rbps, v) },
			"wbps":  func(v string) error { return parseUint64(&m.Wbps, v) },
			"riops": func(v string) error { return parseUint64(&m.Riops, v) },
			"wiops": func(v string) error { return parseUint64(&m.Wiops, v) },
		}); err != nil {
			return ioMax{}, err
		}
	}
	return m, nil
}

// Whether e and other address the same block device (major/minor pair).
func (e ioMax) equal(other ioMax) bool {
	return e.Major == other.Major && e.Minor == other.Minor
}

// Whether e and other have any conflicting non-zero limits.
//
// Same-identity entries with conflicting non-zero limits cannot be merged,
// but they can coexist if one of them has zero limits, which is treated as
// "not set" and overridden by the other. Returns an error if e and other
// share identity with conflicting values.
func (e ioMax) check(other ioMax) error {
	if !e.equal(other) || e == other {
		return nil
	}
	return crex.Wrapf(ErrConflict, "%s %d:%d already set", ioMaxKnob, other.Major, other.Minor)
}

// Leaves e unchanged and always reports no change.
//
// Same-identity entries with conflicting values are rejected upstream by
// check, and identical entries need no merge, so there is nothing to do.
func (e *ioMax) merge(other ioMax) bool {
	return false
}
