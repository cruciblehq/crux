package cgroup

import (
	"strings"

	"github.com/cruciblehq/crux/crex"
)

// Pressure stall information triggers, grouped by resource.
//
// PSI is a kernel utility that reports the fraction of time the cgroup is
// stalled waiting for CPU, memory, or I/O. Each entry specifies a trigger
// that fires when the stall fraction for the corresponding resource exceeds
// the specified threshold for at least the specified duration. The executor
// applies the triggers by writing them to the appropriate cgroup files.
type psi struct {
	CPU    []psiTrigger `knob:"cpu" json:"cpu,omitempty"`       // CPU PSI triggers.
	Memory []psiTrigger `knob:"memory" json:"memory,omitempty"` // Memory PSI triggers.
	IO     []psiTrigger `knob:"io" json:"io,omitempty"`         // I/O PSI triggers.
}

// PSI stall coverage selector.
//
// Distinguishes triggers that fire when at least one task is stalled from
// triggers that fire only when every runnable task is stalled
// simultaneously.
type psiType string

const (
	psiTypeSome psiType = "some" // At least one task stalled.
	psiTypeFull psiType = "full" // All tasks stalled simultaneously.
)

// Parses a PSI stall type name.
func parsePSIType(value string) (psiType, error) {
	s := strings.TrimSpace(value)
	switch psiType(s) {
	case psiTypeSome, psiTypeFull:
		return psiType(s), nil
	default:
		return "", crex.Newf(ErrInvalidGrant, "invalid PSI stall type %q", value)
	}
}

// Merges a PSI trigger list into dst, deduplicating by event type.
func mergePSI(dst *[]psiTrigger, src []psiTrigger, knob string) error {
	for _, tr := range src {
		matched := false
		for i := range *dst {
			e := &(*dst)[i]
			if !e.equal(tr) {
				continue
			}
			matched = true
			if err := e.check(tr, knob); err != nil {
				return err
			}
			e.merge(tr)
			break
		}
		if !matched {
			*dst = append(*dst, tr)
		}
	}
	return nil
}

// Merges PSI triggers into the resource slice selected by knob.
//
// The knob path is the kernel cgroup file name, which selects which slice on
// s.PSI to merge into.
func (s *spec) mergePSITriggers(knob string, triggers []psiTrigger) (bool, error) {
	switch knob {
	case psiCPUKnob:
		return false, mergePSI(&s.PSI.CPU, triggers, knob)
	case psiMemoryKnob:
		return false, mergePSI(&s.PSI.Memory, triggers, knob)
	case psiIOKnob:
		return false, mergePSI(&s.PSI.IO, triggers, knob)
	}
	return false, crex.Newf(ErrUnknownKnob, "unknown PSI knob %q", knob)
}
