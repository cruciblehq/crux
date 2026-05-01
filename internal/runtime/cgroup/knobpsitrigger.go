package cgroup

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/internal/crex"
)

// One PSI threshold trigger within a resource bucket.
//
// A trigger fires when the stall fraction for its coverage type stays above
// the threshold for the configured window length.
type psiTrigger struct {
	Type      psiType `knob:"type" json:"type,omitempty"`           // Stall type for the trigger (some or full).
	Threshold uint64  `knob:"threshold" json:"threshold,omitempty"` // Stall percentage threshold for the trigger (0-100).
	Window    uint64  `knob:"window" json:"window,omitempty"`       // Time window in microseconds for calculating the stall percentage.
}

// Matches a PSI trigger as "type threshold window". Type is any non
// whitespace token (validated separately). Threshold and window are
// unsigned integers. For example, "some 80 1000" would trigger when
// some tasks are stalled on CPU for at least 80% of the time over a
// 1000ms window.
var rePSITrigger = regexp.MustCompile(`^(\S+)\s+(\d+)\s+(\d+)$`)

// Parses a PSI pressure trigger from the textual form "kind threshold window".
func parsePSITrigger(value string) (psiTrigger, error) {
	m := rePSITrigger.FindStringSubmatch(strings.TrimSpace(value))
	if m == nil {
		return psiTrigger{}, crex.Wrapf(ErrInvalidGrant, "psi: expected kind threshold window")
	}
	threshold, err := strconv.ParseUint(m[2], 10, 64)
	if err != nil {
		return psiTrigger{}, crex.Wrap(ErrInvalidGrant, err)
	}
	window, err := strconv.ParseUint(m[3], 10, 64)
	if err != nil {
		return psiTrigger{}, crex.Wrap(ErrInvalidGrant, err)
	}
	pt, err := parsePSIType(m[1])
	if err != nil {
		return psiTrigger{}, err
	}
	return psiTrigger{
		Type:      pt,
		Threshold: threshold,
		Window:    window,
	}, nil
}

// Whether e and other specify the same PSI event type.
func (e psiTrigger) equal(other psiTrigger) bool {
	return e.Type == other.Type
}

// Returns an error if e and other specify the same event type with different parameters.
func (e psiTrigger) check(other psiTrigger, knob string) error {
	if !e.equal(other) || e == other {
		return nil
	}
	return crex.Wrapf(ErrConflict, "%s %s already set", knob, other.Type)
}

// Leaves e unchanged and always reports no change.
//
// Same-identity triggers with conflicting parameters are rejected upstream
// by check, and identical triggers need no merge, so there is nothing to do.
func (e *psiTrigger) merge(other psiTrigger) bool {
	return false
}
