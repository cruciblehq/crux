package cgroup

import "github.com/cruciblehq/crux/crex"

// Per-device I/O cost QoS parameters.
//
// Overrides the io.cost scheduler's quality-of-service targets for one block
// device, controlling whether QoS is active and how aggressively the
// scheduler trades latency against throughput.
type ioCostQoS struct {
	Major  uint32     `knob:"major" json:"major,omitempty"`   // Device major number.
	Minor  uint32     `knob:"minor" json:"minor,omitempty"`   // Device minor number.
	Enable bool       `knob:"enable" json:"enable,omitempty"` // Whether to enable QoS for the device.
	Ctrl   ioCtrlMode `knob:"ctrl" json:"ctrl,omitempty"`     // I/O control mode (auto or user).
	Rpct   float64    `knob:"rpct" json:"rpct,omitempty"`     // Read bandwidth percentage of the device's total bandwidth (0-100).
	Rlat   uint64     `knob:"rlat" json:"rlat,omitempty"`     // Read latency target in microseconds.
	Wpct   float64    `knob:"wpct" json:"wpct,omitempty"`     // Write bandwidth percentage of the device's total bandwidth (0-100).
	Wlat   uint64     `knob:"wlat" json:"wlat,omitempty"`     // Write latency target in microseconds.
	Min    float64    `knob:"min" json:"min,omitempty"`       // Minimum bandwidth percentage of the device's total bandwidth (0-100).
	Max    float64    `knob:"max" json:"max,omitempty"`       // Maximum bandwidth percentage of the device's total bandwidth (0-100).
}

// Parses an io.cost.qos entry.
//
// Expects "major:minor" followed by zero or more space-separated key=value
// pairs drawn from enable, ctrl, rpct, rlat, wpct, wlat, min, and max. The
// device identity is required; QoS fields are optional and omitted ones
// remain at the struct's zero value, which the caller treats as "not set"
// when reconciling against earlier overrides for the same device.
func parseIOCostQoS(value string) (ioCostQoS, error) {
	var q ioCostQoS
	var rest string
	var err error
	q.Major, q.Minor, rest, err = parseMajorMinor(value)
	if err != nil {
		return ioCostQoS{}, err
	}
	if rest != "" {
		if err := parseArgs(rest, map[string]func(string) error{
			"enable": func(v string) error { return parseBool(&q.Enable, v) },
			"ctrl": func(v string) error {
				val, err := parseIOCtrlMode(v)
				if err != nil {
					return err
				}
				q.Ctrl = val
				return nil
			},
			"rpct": func(v string) error { return parseFloat64(&q.Rpct, v) },
			"rlat": func(v string) error { return parseUint64(&q.Rlat, v) },
			"wpct": func(v string) error { return parseFloat64(&q.Wpct, v) },
			"wlat": func(v string) error { return parseUint64(&q.Wlat, v) },
			"min":  func(v string) error { return parseFloat64(&q.Min, v) },
			"max":  func(v string) error { return parseFloat64(&q.Max, v) },
		}); err != nil {
			return ioCostQoS{}, err
		}
	}
	return q, nil
}

// Whether e and other address the same block device (major/minor pair).
func (e ioCostQoS) equal(other ioCostQoS) bool {
	return e.Major == other.Major && e.Minor == other.Minor
}

// Returns an error if e and other share identity with conflicting values.
func (e ioCostQoS) check(other ioCostQoS) error {
	if !e.equal(other) || e == other {
		return nil
	}
	return crex.Newf(ErrConflict, "%s %d:%d already set", ioCostQoSKnob, other.Major, other.Minor)
}

// Leaves e unchanged and always reports no change.
//
// Same-identity entries with conflicting values are rejected upstream by
// check, and identical entries need no merge, so there is nothing to do.
func (e *ioCostQoS) merge(other ioCostQoS) bool {
	return false
}
