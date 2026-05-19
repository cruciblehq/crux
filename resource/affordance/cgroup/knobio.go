package cgroup

import (
	"strings"

	"github.com/cruciblehq/crux/crex"
)

// Block I/O control model.
//
// Groups the cgroup v2 io controller knobs covering group-level priority and
// weight as well as per-device overrides for weight, bandwidth and IOPS
// caps, latency targets, and the cost-based scheduler.
type io struct {
	PrioClass     ioPrioClass      `knob:"prio.class" default:"idle" json:"prioClass,omitempty"` // I/O priority class (idle = served only when no other I/O is pending).
	Weight        uint16           `knob:"weight" default:"1" json:"weight,omitempty"`           // Default I/O weight for the cgroup.
	WeightDevices []ioWeightDevice `knob:"weight.devices" json:"weightDevices,omitempty"`        // Per-device I/O weight overrides.
	Max           []ioMax          `knob:"max" json:"max,omitempty"`                             // Per-device I/O bandwidth and IOPS caps.
	Latency       []ioLatency      `knob:"latency" json:"latency,omitempty"`                     // Per-device I/O latency targets.
	Cost          []ioCost         `knob:"cost.model" json:"cost,omitempty"`                     // Per-device I/O cost model coefficients.
	CostQoS       []ioCostQoS      `knob:"cost.qos" json:"costQos,omitempty"`                    // Per-device I/O cost QoS parameters.
}

// I/O scheduling priority class within the block layer.
//
// Higher classes are dispatched ahead of lower ones regardless of weight;
// within a single class, requests are scheduled proportionally to weight.
type ioPrioClass string

const (
	ioPrioClassRT   ioPrioClass = "rt"   // Real-time (highest priority, deadline scheduling).
	ioPrioClassBE   ioPrioClass = "be"   // Best-effort (default class, weight-based scheduling).
	ioPrioClassIdle ioPrioClass = "idle" // Idle (served only when no other I/O is pending).
)

// Parses an I/O priority class name.
func parseIOPrioClass(value string) (ioPrioClass, error) {
	s := strings.TrimSpace(value)
	switch ioPrioClass(s) {
	case ioPrioClassRT, ioPrioClassBE, ioPrioClassIdle:
		return ioPrioClass(s), nil
	default:
		return "", crex.Wrapf(ErrInvalidGrant, "invalid I/O priority class %q", value)
	}
}

// I/O cost controller mode.
//
// Selects whether the kernel measures device characteristics on its own or
// trusts user-supplied cost-model coefficients verbatim.
type ioCtrlMode string

const (
	ioCtrlModeAuto ioCtrlMode = "auto" // Kernel auto-detects device capabilities.
	ioCtrlModeUser ioCtrlMode = "user" // User supplies cost model coefficients via io.cost.model.
)

// Parses an I/O cost controller mode name.
func parseIOCtrlMode(value string) (ioCtrlMode, error) {
	s := strings.TrimSpace(value)
	switch ioCtrlMode(s) {
	case ioCtrlModeAuto, ioCtrlModeUser:
		return ioCtrlMode(s), nil
	default:
		return "", crex.Wrapf(ErrInvalidGrant, "invalid I/O cost controller mode %q", value)
	}
}

// Parses an io.weight scalar value from pre-split fields.
//
// Accepts "weight" or "default weight" forms. Returns the parsed weight.
func parseIOWeightScalar(parts []string) (uint16, error) {
	var w uint16
	weight := parts[0]
	if weight == "default" {
		if len(parts) < 2 {
			return 0, crex.Wrapf(ErrInvalidGrant, "value required after 'default' in %q", ioWeightKnob)
		}
		weight = parts[1]
	}
	if err := parseUint16(&w, weight); err != nil {
		return 0, err
	}
	return w, nil
}
