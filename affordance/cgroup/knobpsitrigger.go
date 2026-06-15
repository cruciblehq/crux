package cgroup

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/crex"
)

// One PSI threshold trigger within a resource bucket.
//
// A trigger fires when the stall fraction for its coverage type stays above
// the threshold for the configured window length.
type psiTrigger struct {
	Type      psiType `knob:"type" json:"type,omitempty"`           // Stall type for the trigger (some or full).
	Threshold uint64  `knob:"threshold" json:"threshold,omitempty"` // Cumulative stall time in microseconds within Window that arms the trigger.
	Window    uint64  `knob:"window" json:"window,omitempty"`       // Tracking window in microseconds over which Threshold is accumulated.
}

// Matches a PSI trigger as "type threshold window". Type is any non
// whitespace token (validated separately). Threshold and window are
// unsigned integers in microseconds. For example, "some 150000 1000000"
// fires when at least one task is stalled on CPU for a cumulative 150ms
// within any 1s window.
var rePSITrigger = regexp.MustCompile(`^(\S+)\s+(\d+)\s+(\d+)$`)

// Parses a PSI pressure trigger from the textual form "type threshold window".
func parsePSITrigger(value string) (psiTrigger, error) {
	m := rePSITrigger.FindStringSubmatch(strings.TrimSpace(value))
	if m == nil {
		return psiTrigger{}, crex.Newf(ErrInvalidGrant, "expected psi trigger as type threshold window")
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
	return crex.Newf(ErrConflict, "%s %s already set", knob, other.Type)
}

// Leaves e unchanged and always reports no change.
//
// Same-identity triggers with conflicting parameters are rejected upstream
// by check, and identical triggers need no merge, so there is nothing to do.
func (e *psiTrigger) merge(other psiTrigger) bool {
	return false
}
