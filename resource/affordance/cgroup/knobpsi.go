package cgroup

import (
	"strings"

	"github.com/cruciblehq/crux/crex"
)

// Pressure stall information triggers, grouped by resource.
//
// PSI is a kernel facility that reports the fraction of wall time the
// cgroup is stalled waiting for CPU, memory, I/O, or IRQ servicing. Each
// resource bucket holds an independent set of triggers that fire when their
// stall threshold is sustained over the configured window.
type psi struct {
	CPU    []psiTrigger `knob:"cpu" json:"cpu,omitempty"`       // CPU PSI triggers.
	Memory []psiTrigger `knob:"memory" json:"memory,omitempty"` // Memory PSI triggers.
	IO     []psiTrigger `knob:"io" json:"io,omitempty"`         // I/O PSI triggers.
	IRQ    []psiTrigger `knob:"irq" json:"irq,omitempty"`       // IRQ PSI triggers.
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
		return "", crex.Wrapf(ErrInvalidGrant, "invalid PSI stall type %q", value)
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
// The knob path encodes the target resource ("psi.cpu", "psi.memory", "psi.io",
// "psi.irq"), which selects which slice on s.PSI to merge into.
func (s *spec) mergePSITriggers(knob string, triggers []psiTrigger) (bool, error) {
	switch knob {
	case psiCPUKnob:
		return false, mergePSI(&s.PSI.CPU, triggers, knob)
	case psiMemoryKnob:
		return false, mergePSI(&s.PSI.Memory, triggers, knob)
	case psiIOKnob:
		return false, mergePSI(&s.PSI.IO, triggers, knob)
	case psiIRQKnob:
		return false, mergePSI(&s.PSI.IRQ, triggers, knob)
	}
	return false, crex.Wrapf(ErrUnknownKnob, "unknown PSI knob %q", knob)
}
