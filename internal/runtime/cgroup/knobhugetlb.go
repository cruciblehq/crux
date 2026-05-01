package cgroup

import (
	"strings"

	"github.com/cruciblehq/crux/internal/crex"
)

// Huge page limit for a single page size.
//
// Huge pages are a Linux memory management feature that allows allocations
// to use page sizes larger than the default 4 KiB; the hugetlb controller
// bounds how many pages of each size the cgroup may consume.
type hugeTLB struct {
	Size    string `knob:"size" json:"size,omitempty"`       // Kernel-format page size string (for example "2MB" or "1GB").
	Max     uint64 `knob:"max" json:"max,omitempty"`         // Cap on in-use pages of this size; zero means unlimited.
	RsvdMax uint64 `knob:"rsvdMax" json:"rsvdMax,omitempty"` // Cap on reserved pages of this size; zero means unlimited.
}

// Parses a hugetlb entry.
//
// The value is "size [key=value...]". Size is the page size as exposed by
// the kernel (for example "2MB" or "1GB") and is required. Accepted keys are
// max (page count cap) and rsvd_max (reservation cap); both are optional and
// default to zero, which the controller treats as unlimited.
func parseHugeTLB(value string) (hugeTLB, error) {
	size, rest, _ := strings.Cut(strings.TrimSpace(value), " ")
	if size == "" {
		return hugeTLB{}, crex.Wrapf(ErrInvalidGrant, "page size required")
	}
	h := hugeTLB{Size: size}
	if rest != "" {
		if err := parseArgs(rest, map[string]func(string) error{
			"max":      func(v string) error { return parseUint64(&h.Max, v) },
			"rsvd_max": func(v string) error { return parseUint64(&h.RsvdMax, v) },
		}); err != nil {
			return hugeTLB{}, err
		}
	}
	return h, nil
}

// Whether e and other specify the same huge page size.
func (e hugeTLB) equal(other hugeTLB) bool {
	return e.Size == other.Size
}

// Returns ErrConflict when e and other share Size but differ in any limit field.
//
// Identical entries (e == other) are accepted as idempotent repeats and entries
// for different page sizes are unrelated; everything else is a conflict.
func (e hugeTLB) check(other hugeTLB) error {
	if !e.equal(other) || e == other {
		return nil
	}
	return crex.Wrapf(ErrConflict, "%s %s already set", "hugetlb", other.Size)
}

// Leaves e unchanged and always reports no change.
//
// Same-identity entries with conflicting values are rejected upstream by
// check, and identical entries need no merge, so there is nothing to do.
func (e *hugeTLB) merge(other hugeTLB) bool {
	return false
}
