package cgroup

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/internal/crex"
)

// Per-device I/O weight override.
//
// Overrides the cgroup's default proportional I/O weight for one block
// device. Weights interact with io.max so that a per-device weight
// determines the share within the device's available bandwidth and io.max
// then caps that share.
type ioWeightDevice struct {
	Major  uint32 `knob:"major" json:"major,omitempty"`   // Device major number.
	Minor  uint32 `knob:"minor" json:"minor,omitempty"`   // Device minor number.
	Weight uint16 `knob:"weight" json:"weight,omitempty"` // I/O weight for the device (1-10000).
}

// Matches an io.weight per-device entry as "major:minor weight".
// Major, minor, and weight are unsigned integers. For example, "8:0 500"
// specifies major 8 and minor 0 with a weight of 500.
var reIOWeightDevice = regexp.MustCompile(`^(\d+):(\d+)(?:\s+(\S+))?$`)

// Parses an io.weight per-device entry.
//
// Expects "major:minor weight". The kernel also accepts removal forms
// ("major:minor" and "major:minor default") that clear a prior override,
// but the spec only ever adds positive overrides; a removal in a grant
// would either contradict an earlier set or undo state the spec never
// owned, so removals are rejected as ErrInvalidGrant.
func parseIOWeightDevice(value string) (ioWeightDevice, error) {
	m := reIOWeightDevice.FindStringSubmatch(strings.TrimSpace(value))
	if m == nil {
		return ioWeightDevice{}, crex.Wrapf(ErrInvalidGrant, "expected major:minor weight for io.weight device entry")
	}
	maj, err := strconv.ParseUint(m[1], 10, 32)
	if err != nil {
		return ioWeightDevice{}, crex.Wrap(ErrInvalidGrant, err)
	}
	min, err := strconv.ParseUint(m[2], 10, 32)
	if err != nil {
		return ioWeightDevice{}, crex.Wrap(ErrInvalidGrant, err)
	}
	if m[3] == "" || m[3] == "default" {
		return ioWeightDevice{}, crex.Wrapf(ErrInvalidGrant, "restrictive form %q not allowed", value)
	}
	wd := ioWeightDevice{Major: uint32(maj), Minor: uint32(min)}
	if err := parseUint16(&wd.Weight, m[3]); err != nil {
		return ioWeightDevice{}, err
	}
	return wd, nil
}

// Whether e and other address the same block device (major/minor pair).
func (e ioWeightDevice) equal(other ioWeightDevice) bool {
	return e.Major == other.Major && e.Minor == other.Minor
}

// Returns an error if e and other share identity with conflicting values.
func (e ioWeightDevice) check(other ioWeightDevice) error {
	if !e.equal(other) || e == other {
		return nil
	}
	return crex.Wrapf(ErrConflict, "%s %d:%d already set", ioWeightKnob, other.Major, other.Minor)
}

// Leaves e unchanged and always reports no change.
//
// Same-identity entries with conflicting values are rejected upstream by
// check, and identical entries need no merge, so there is nothing to do.
func (e *ioWeightDevice) merge(other ioWeightDevice) bool {
	return false
}
