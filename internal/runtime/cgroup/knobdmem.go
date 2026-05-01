package cgroup

import (
	"strings"

	"github.com/cruciblehq/crux/internal/crex"
)

// Device memory limit for one GPU or accelerator memory region.
//
// Device memory regions are vendor-defined named partitions of an accelerator's
// onboard memory; the dmem controller bounds how much of each region the cgroup
// may consume.
type dmem struct {
	Region string `knob:"region" json:"region,omitempty"` // Device memory region identifier (e.g., "gpu0").
	Max    uint64 `knob:"max" json:"max,omitempty"`       // Maximum device memory usage for the region in bytes.
	Min    uint64 `knob:"min" json:"min,omitempty"`       // Minimum device memory usage for the region in bytes.
	Low    uint64 `knob:"low" json:"low,omitempty"`       // Low device memory limit for the region in bytes.
}

// Parses a dmem limit entry of the form "region [value]".
//
// setField is called with the parsed uint64 value (or 0 when absent) to
// populate the appropriate limit field on the returned Dmem entry.
func parseDmemValue(value string, setField func(entry *dmem, v uint64)) (dmem, error) {
	region, rest, _ := strings.Cut(strings.TrimSpace(value), " ")
	if region == "" {
		return dmem{}, crex.Wrapf(ErrInvalidGrant, "device memory region required")
	}
	entry := dmem{Region: region}
	if rest == "" {
		setField(&entry, 0)
		return entry, nil
	}
	var parsed uint64
	if err := parseUint64(&parsed, rest); err != nil {
		return dmem{}, err
	}
	setField(&entry, parsed)
	return entry, nil
}

// Parses a dmem entry, routing the parsed value to the field selected by knob.
//
// The knob tag on the []dmem field is "dmem", but individual grants use
// "dmem.max", "dmem.min", or "dmem.low" to specify which limit to set on the
// entry. The trailing segment of the knob path selects the target field.
func parseDmemEntry(knob string, value string) (dmem, error) {
	switch knob {
	case dmemMaxKnob:
		return parseDmemValue(value, func(e *dmem, v uint64) { e.Max = v })
	case dmemMinKnob:
		return parseDmemValue(value, func(e *dmem, v uint64) { e.Min = v })
	case dmemLowKnob:
		return parseDmemValue(value, func(e *dmem, v uint64) { e.Low = v })
	}
	return dmem{}, crex.Wrapf(ErrUnknownKnob, "unknown dmem knob %q", knob)
}

// Whether e and other constrain the same device memory region.
func (e dmem) equal(other dmem) bool {
	return e.Region == other.Region
}

// Checks for conflicting limits between e and other.
//
// Returns an error if e and other share identity with conflicting per-field
// limit values.
func (e dmem) check(other dmem) error {
	alreadySetErr := "%s %s already set"

	if !e.equal(other) {
		return nil
	}
	if e.Max != 0 && other.Max != 0 && e.Max != other.Max {
		return crex.Wrapf(ErrConflict, alreadySetErr, dmemMaxKnob, other.Region)
	}
	if e.Min != 0 && other.Min != 0 && e.Min != other.Min {
		return crex.Wrapf(ErrConflict, alreadySetErr, dmemMinKnob, other.Region)
	}
	if e.Low != 0 && other.Low != 0 && e.Low != other.Low {
		return crex.Wrapf(ErrConflict, alreadySetErr, dmemLowKnob, other.Region)
	}
	return nil
}

// Merges e and other by unioning non-zero limit values.
//
// Returns true if e was modified by the merge, or false if e already contained
// all non-zero values in other. For example, if e has Max=10 and other has
// Max=0, Min=5, and Low=15, e would be updated to Max=10, Min=5, Low=15 and
// the method would return true. If e had already been Max=10, Min=5, Low=15,
// the method would return false since no change was needed.
func (e *dmem) merge(other dmem) bool {
	if !e.equal(other) {
		return false
	}
	before := *e
	if e.Max == 0 {
		e.Max = other.Max
	}
	if e.Min == 0 {
		e.Min = other.Min
	}
	if e.Low == 0 {
		e.Low = other.Low
	}
	return *e != before
}
