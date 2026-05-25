package manifest

// VM image builder requirements compiled from affordance grants.
//
// Each affordance that needs VM-level support contributes entries here. The VM
// image builder validates KernelFeatures against the selected kernel's .config
// and fails the build if any flag is absent. Init applies Sysctls and Nftables
// at boot, before cruxd starts. A change to KernelFeatures or Sysctls requires
// a VM image rebuild; a change to Nftables alone only requires regenerating the
// init configuration.
type VMSpec struct {

	// Linux kernel CONFIG_* flags the VM must include.
	//
	// Each entry is a flag name without the CONFIG_ prefix, for example
	// "NETFILTER" or "FUSE_FS". The VM image builder validates each flag
	// against the selected kernel's .config and fails if any is missing.
	KernelFeatures []string `codec:"kernel_features,omitempty"`

	// Boot-time sysctl relaxations applied over the hardened baseline.
	//
	// Each entry overrides one sysctl set by the hardened boot configuration.
	// An affordance relaxes a sysctl when the baseline is more restrictive
	// than the workload requires.
	Sysctls map[string]string `codec:"sysctls,omitempty"`

	// Init-time nftables rules punched into the VM-level deny-all firewall.
	//
	// The VM starts with a deny-all nftables ruleset. Affordances that require
	// network access contribute rules here to open only the necessary paths.
	Nftables []VMNftRule `codec:"nftables,omitempty"`
}

// Single nftables rule applied to the VM-level deny-all firewall.
type VMNftRule struct {

	// nftables table name.
	Table string `codec:"table"`

	// nftables chain name.
	Chain string `codec:"chain"`

	// nftables rule expression.
	Rule string `codec:"rule"`
}
