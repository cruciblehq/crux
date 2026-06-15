package kernel

import (
	"github.com/cruciblehq/crux/crex"
)

// Kernel requirements compiled from .kernel grants.
//
// Each affordance that needs kernel-level support contributes entries here. The
// orchestrator validates all fields against the selected kernel at provisioning
// time and fails the deployment if any requirement is unmet. A change to any
// field requires a VM image rebuild.
type Spec struct {

	// Linux kernel CONFIG_* flags the kernel must include.
	//
	// Each entry is a flag name without the CONFIG_ prefix, for example
	// "NETFILTER" or "FUSE_FS". The orchestrator validates each flag against
	// the selected kernel's .config at provisioning time.
	Features []string `codec:"features,omitempty"`

	// Kernel modules that must be available, either built-in or loadable.
	//
	// Each entry is a module name, for example "fuse" or "dm_crypt". The
	// orchestrator verifies availability via modinfo or the module list.
	Modules []string `codec:"modules,omitempty"`

	// Minimum kernel versions the kernel must meet.
	//
	// Each entry is a version string, for example "5.15" or "6.1". The
	// orchestrator verifies the running kernel satisfies all entries.
	Versions []string `codec:"versions,omitempty"`

	// Boot parameters that must appear in /proc/cmdline.
	//
	// Each entry is a parameter token, for example "mitigations=off" or
	// "iommu=pt". The orchestrator verifies presence at provisioning time.
	BootParams []string `codec:"boot_params,omitempty"`

	// Linux Security Modules that must be active.
	//
	// Each entry is an LSM name, for example "apparmor" or "bpf". The
	// orchestrator checks /sys/kernel/security/lsm at provisioning time.
	LSMs []string `codec:"lsms,omitempty"`

	// CPU hardware feature flags the kernel must expose.
	//
	// Each entry is a feature flag from /proc/cpuinfo, for example "sgx"
	// or "avx512f". The orchestrator validates availability at provisioning time.
	HWFeatures []string `codec:"hw_features,omitempty"`
}

// Validates the kernel spec.
//
// All slice entries must be non-empty.
func (s *Spec) Validate() error {
	for _, group := range []struct {
		label string
		items []string
	}{
		{"feature", s.Features},
		{"module", s.Modules},
		{"version", s.Versions},
		{"boot param", s.BootParams},
		{"LSM", s.LSMs},
		{"hw feature", s.HWFeatures},
	} {
		if err := validateNonEmpty(group.label, group.items); err != nil {
			return err
		}
	}
	return nil
}

// Validates that no entry in items is empty, naming the offending index with
// label in the error.
func validateNonEmpty(label string, items []string) error {
	for i, s := range items {
		if s == "" {
			return crex.Wrapf(ErrInvalidSpec, "empty %s at index %d", label, i)
		}
	}
	return nil
}

// Merges src into the receiver using union semantics.
//
// All slice fields are deduplicated by value.
func (s *Spec) Merge(src Spec) {
	s.Features = mergeStrings(s.Features, src.Features)
	s.Modules = mergeStrings(s.Modules, src.Modules)
	s.Versions = mergeStrings(s.Versions, src.Versions)
	s.BootParams = mergeStrings(s.BootParams, src.BootParams)
	s.LSMs = mergeStrings(s.LSMs, src.LSMs)
	s.HWFeatures = mergeStrings(s.HWFeatures, src.HWFeatures)
}

// Appends elements of src not already in dst, deduplicating by value.
func mergeStrings(dst, src []string) []string {
	seen := make(map[string]struct{}, len(dst))
	for _, s := range dst {
		seen[s] = struct{}{}
	}
	for _, s := range src {
		if _, dup := seen[s]; !dup {
			dst = append(dst, s)
			seen[s] = struct{}{}
		}
	}
	return dst
}
