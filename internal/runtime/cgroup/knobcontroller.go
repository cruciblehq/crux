package cgroup

import (
	"strings"

	"github.com/cruciblehq/crux/internal/crex"
)

// controller name in cgroup.subtree_control.
//
// These are the only controllers that can be delegated to child cgroups and
// thus the only controllers that can be specified in cgroup.subtree_control.
// Other controllers can be configured for the current cgroup but not delegated
// to children.
type controller string

const (
	controllerCPU     controller = "cpu"     // CPU bandwidth and scheduling priority.
	controllerCPUSet  controller = "cpuset"  // CPU and memory node affinity.
	controllerIO      controller = "io"      // Block I/O weight, priority, and per-device limits.
	controllerMemory  controller = "memory"  // Memory limits and reclaim protection.
	controllerHugeTLB controller = "hugetlb" // Huge page limits per page size.
	controllerPids    controller = "pids"    // Process and thread count limit.
	controllerRDMA    controller = "rdma"    // RDMA resource limits per HCA device.
	controllerMisc    controller = "misc"    // Miscellaneous scalar resource limits.
	controllerDevMem  controller = "dmem"    // Device memory limits per region.
)

// Parses a cgroup controller name.
//
// Used for parsing entries in cgroup.subtree_control, which accepts a list of
// controllers to delegate to child cgroups. The list must be non-empty and
// contain no duplicates. Controller names must not include the +/- prefix used
// in the kernel interface; all entries are treated as additions.
func parseController(value string) (controller, error) {
	s := strings.TrimSpace(value)
	switch controller(s) {
	case controllerCPU, controllerCPUSet, controllerIO, controllerMemory,
		controllerHugeTLB, controllerPids, controllerRDMA, controllerMisc,
		controllerDevMem:
		return controller(s), nil
	default:
		return "", crex.Wrapf(ErrInvalidGrant, "invalid cgroup controller %q", value)
	}
}

// Parses a space-separated list of cgroup controller names.
//
// Used for the cgroup.subtree_control, which accepts a list of controllers
// to delegate to child cgroups. The list must be non-empty and contain no
// duplicates. Controller names must not include the +/- prefix used in the
// kernel interface; all entries are treated as additions.
func parseSubtreeControl(value string) ([]controller, error) {
	tokens := strings.Fields(value)
	if len(tokens) == 0 {
		return nil, crex.Wrapf(ErrInvalidGrant, "at least one controller required for %q", subtreeControlKnob)
	}
	controllers := make([]controller, 0, len(tokens))
	seen := make(map[controller]struct{}, len(tokens))
	for _, token := range tokens {
		if strings.HasPrefix(token, "+") || strings.HasPrefix(token, "-") {
			return nil, crex.Wrapf(ErrInvalidGrant, "%q controller names must not include +/- prefix as in %q", subtreeControlKnob, token)
		}
		c, err := parseController(token)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[c]; ok {
			return nil, crex.Wrapf(ErrInvalidGrant, "duplicate controller %q for %q", token, subtreeControlKnob)
		}
		seen[c] = struct{}{}
		controllers = append(controllers, c)
	}
	return controllers, nil
}
