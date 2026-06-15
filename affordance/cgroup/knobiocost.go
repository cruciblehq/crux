package cgroup

import "github.com/cruciblehq/crux/crex"

// Per-device I/O cost model coefficients.
//
// Overrides the kernel's auto-detected linear cost model for one block device.
// The io.cost scheduler converts every request into a cost using these
// coefficients (which describe the device's expected throughput) and charges
// that cost against the cgroup's share of the device, throttling cgroups that
// exceed their quality-of-service target. A zero coefficient leaves the
// kernel's auto-detected value in effect for that dimension.
type ioCost struct {
	Major     uint32 `knob:"major" json:"major,omitempty"`         // Device major number.
	Minor     uint32 `knob:"minor" json:"minor,omitempty"`         // Device minor number.
	Rbps      uint64 `knob:"rbps" json:"rbps,omitempty"`           // Sequential read throughput coefficient in bytes per second.
	Rseqiops  uint64 `knob:"rseqiops" json:"rseqiops,omitempty"`   // Sequential read IOPS coefficient.
	Rrandiops uint64 `knob:"rrandiops" json:"rrandiops,omitempty"` // Random read IOPS coefficient.
	Wbps      uint64 `knob:"wbps" json:"wbps,omitempty"`           // Sequential write throughput coefficient in bytes per second.
	Wseqiops  uint64 `knob:"wseqiops" json:"wseqiops,omitempty"`   // Sequential write IOPS coefficient.
	Wrandiops uint64 `knob:"wrandiops" json:"wrandiops,omitempty"` // Random write IOPS coefficient.
}

// Parses an io.cost.model entry.
//
// The value is "major:minor [key=value...]". Major and minor identify the
// block device. Optional keys are rbps, rseqiops, rrandiops, wbps, wseqiops,
// wrandiops; each sets one coefficient of the linear cost model. Omitted
// keys default to zero, leaving that dimension on the kernel's auto-detected
// value.
func parseIOCost(value string) (ioCost, error) {
	var c ioCost
	var rest string
	var err error
	c.Major, c.Minor, rest, err = parseMajorMinor(value)
	if err != nil {
		return ioCost{}, err
	}
	if rest != "" {
		if err := parseArgs(rest, map[string]func(string) error{
			"rbps":      func(v string) error { return parseUint64(&c.Rbps, v) },
			"wbps":      func(v string) error { return parseUint64(&c.Wbps, v) },
			"rseqiops":  func(v string) error { return parseUint64(&c.Rseqiops, v) },
			"rrandiops": func(v string) error { return parseUint64(&c.Rrandiops, v) },
			"wseqiops":  func(v string) error { return parseUint64(&c.Wseqiops, v) },
			"wrandiops": func(v string) error { return parseUint64(&c.Wrandiops, v) },
		}); err != nil {
			return ioCost{}, err
		}
	}
	return c, nil
}

// Whether e and other address the same block device (major/minor pair).
func (e ioCost) equal(other ioCost) bool {
	return e.Major == other.Major && e.Minor == other.Minor
}

// Returns ErrConflict when e and other target the same device but disagree on any coefficient.
//
// Identical entries (e == other) are accepted as idempotent repeats; entries
// for different devices are unrelated. Per-coefficient union is not attempted
// because the cost model is calibrated as a whole and partial overrides would
// silently distort the kernel's auto-detected baseline.
func (e ioCost) check(other ioCost) error {
	if !e.equal(other) || e == other {
		return nil
	}
	return crex.Newf(ErrConflict, "%s %d:%d already set", ioCostModelKnob, other.Major, other.Minor)
}

// Leaves e unchanged and always reports no change.
//
// Same-identity entries with conflicting values are rejected upstream by
// check, and identical entries need no merge, so there is nothing to do.
func (e *ioCost) merge(other ioCost) bool {
	return false
}
