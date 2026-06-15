package cgroup

import "github.com/cruciblehq/crux/crex"

// Per-device I/O latency target.
//
// Asks the kernel to keep request latency on one block device below the
// specified target by throttling cgroups whose I/O is making the device
// exceed it.
type ioLatency struct {
	Major  uint32 `knob:"major" json:"major,omitempty"`   // Device major number.
	Minor  uint32 `knob:"minor" json:"minor,omitempty"`   // Device minor number.
	Target uint64 `knob:"target" json:"target,omitempty"` // Latency target in microseconds.
}

// Parses an io.latency entry.
//
// Expects "major:minor" optionally followed by "target=<microseconds>". The
// device identity is required; an omitted target leaves the field at zero,
// which the caller treats as "not set" when reconciling against earlier
// overrides for the same device.
func parseIOLatency(value string) (ioLatency, error) {
	var l ioLatency
	var rest string
	var err error
	l.Major, l.Minor, rest, err = parseMajorMinor(value)
	if err != nil {
		return ioLatency{}, err
	}
	if rest != "" {
		if err := parseArgs(rest, map[string]func(string) error{
			"target": func(v string) error { return parseUint64(&l.Target, v) },
		}); err != nil {
			return ioLatency{}, err
		}
	}
	return l, nil
}

// Whether e and other address the same block device (major/minor pair).
func (e ioLatency) equal(other ioLatency) bool {
	return e.Major == other.Major && e.Minor == other.Minor
}

// Returns an error if e and other share identity with conflicting values.
func (e ioLatency) check(other ioLatency) error {
	if !e.equal(other) || e == other {
		return nil
	}
	return crex.Newf(ErrConflict, "%s %d:%d already set", ioLatencyKnob, other.Major, other.Minor)
}

// Leaves e unchanged and always reports no change.
//
// Same-identity entries with conflicting values are rejected upstream by
// check, and identical entries need no merge, so there is nothing to do.
func (e *ioLatency) merge(other ioLatency) bool {
	return false
}
