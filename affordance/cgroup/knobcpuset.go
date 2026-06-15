package cgroup

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/crex"
)

// CPU and memory node affinity for the cgroup.
//
// The cpuset controller restricts which CPUs and NUMA nodes the cgroup may
// schedule on, optionally carving out exclusive partitions.
type cpuSet struct {
	CPUs      indexList `knob:"cpus" json:"cpus,omitempty"`                            // CPUs allowed for the cgroup (e.g. "0-3,5").
	Mems      indexList `knob:"mems" json:"mems,omitempty"`                            // Memory nodes allowed for the cgroup (e.g. "0-1").
	Partition partition `knob:"partition" default:"member" json:"partition,omitempty"` // Partitioning mode for the CPUs.
	Exclusive indexList `knob:"exclusive" json:"exclusive,omitempty"`                  // CPUs reserved exclusively for this cgroup.
}

// Partition mode for cpuset CPU isolation.
//
// Determines how the partition formed by the cgroup's CPUs relates to its
// parent and to the scheduler's load balancer.
type partition string

const (
	partitionMember   partition = "member"   // Non-isolated, shares parent's CPUs.
	partitionRoot     partition = "root"     // Partition root, owns its CPUs exclusively.
	partitionIsolated partition = "isolated" // Like root, but also removed from the scheduler's load-balancing.
)

// Parses a cpuset partition mode name.
func parseCPUSetPartition(value string) (partition, error) {
	s := strings.TrimSpace(value)
	switch partition(s) {
	case partitionMember, partitionRoot, partitionIsolated:
		return partition(s), nil
	default:
		return "", crex.Newf(ErrInvalidGrant, "invalid cpuset partition mode %q", value)
	}
}

// Single index or contiguous range within a cpulist or nodelist.
//
// Represents either a single index (Start == End) or a contiguous range of
// indices (Start < End) within a cpulist or nodelist.
type indexRange struct {
	Start uint32 `json:"start,omitempty"` // First index in the range (inclusive).
	End   uint32 `json:"end,omitempty"`   // Last index in the range (inclusive); equals Start for a single index.
}

// Parsed cpulist or nodelist as a sequence of index ranges.
//
// Corresponds to the kernel's standard list format: comma-separated decimal
// indices and ranges (e.g. "0-3,5"). Used for both CPU affinity (cpuset.cpus)
// and NUMA node affinity (cpuset.mems) knobs, which share identical syntax.
// Values are local indices into the corresponding allocation, not kernel-frame
// identifiers. Index 0 means "the first CPU (or memory node) in the allocation",
// which the composer translates to a kernel CPU ID at deployment time using the
// allocation's binding map. An empty list means "no per-cgroup affinity set"
// and the kernel inherits the parent's effective set per cgroupv2 semantics.
type indexList []indexRange

// Matches a single CPU/memory index list token.
//
// A token is either a bare integer (e.g., "3") or an inclusive integer range
// (e.g., "0-3"). Surrounding whitespace is not allowed.
var reIndexToken = regexp.MustCompile(`^(\d+)(?:-(\d+))?$`)

// Parses a comma-separated index list into dst.
//
// The value is a comma-separated list of indices and inclusive index ranges
// (e.g. "0-3,5"). The list must be non-empty and contain no malformed tokens or
// descending ranges. Returns an error if the value is invalid. On success, dst
// is set to the normalized list of index ranges corresponding to the value.
func parseIndexList(dst *indexList, value string) error {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return crex.Newf(ErrInvalidGrant, "list required")
	}

	tokens := strings.Split(raw, ",")
	ranges := make(indexList, 0, len(tokens))
	for _, token := range tokens {
		r, err := parseIndexToken(token)
		if err != nil {
			return err
		}
		ranges = append(ranges, r)
	}

	*dst = normalizeIndexList(ranges)
	return nil
}

// Parses a single comma-separated token from an index list.
//
// A token is either a bare non-negative integer ("3") or an inclusive range
// of non-negative integers ("0-3"). No surrounding whitespace is permitted.
// Returns an error if the token is empty, malformed, or specifies a
// descending range.
func parseIndexToken(token string) (indexRange, error) {
	m := reIndexToken.FindStringSubmatch(token)
	if m == nil {
		return indexRange{}, crex.Newf(ErrInvalidGrant, "invalid list element %q", token)
	}
	start, err := strconv.ParseUint(m[1], 10, 32)
	if err != nil {
		return indexRange{}, crex.Wrap(ErrInvalidGrant, err)
	}
	if m[2] == "" {
		n := uint32(start)
		return indexRange{Start: n, End: n}, nil
	}
	end, err := strconv.ParseUint(m[2], 10, 32)
	if err != nil {
		return indexRange{}, crex.Wrap(ErrInvalidGrant, err)
	}
	if start > end {
		return indexRange{}, crex.Newf(ErrInvalidGrant, "invalid descending range %q", token)
	}
	return indexRange{Start: uint32(start), End: uint32(end)}, nil
}

// Normalizes a list of index ranges.
//
// The ranges are sorted by start index, then merged if they overlap or are
// adjacent. The result is a normalized list of non-overlapping, non-adjacent
// ranges. For example, "0-2,4,5,7-9" would be normalized to "0-2,4-5,7-9"
// and "0-3,2-5" would be normalized to "0-5". This ensures a canonical form
// for semantically equivalent index lists.
func normalizeIndexList(ranges indexList) indexList {
	if len(ranges) <= 1 {
		return ranges
	}
	slices.SortFunc(ranges, func(a, b indexRange) int {
		if a.Start < b.Start {
			return -1
		}
		if a.Start > b.Start {
			return 1
		}
		return 0
	})
	merged := ranges[:1]
	for _, r := range ranges[1:] {
		last := &merged[len(merged)-1]
		if r.Start <= last.End+1 {
			if r.End > last.End {
				last.End = r.End
			}
		} else {
			merged = append(merged, r)
		}
	}
	return merged
}
