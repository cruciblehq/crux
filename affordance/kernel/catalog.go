package kernel

// Requirement types accepted as the first positional argument of a .kernel grant.
const (
	typeConfig  = "config"  // CONFIG_* build flag, without the CONFIG_ prefix.
	typeModule  = "module"  // Loadable kernel module name.
	typeVersion = "version" // Minimum kernel version.
	typeBoot    = "boot"    // Boot command-line parameter token.
	typeLSM     = "lsm"     // Linux Security Module name.
	typeHW      = "hw"      // CPU hardware feature flag.
)

// Positions of the positional arguments in a kernel grant.
const (
	argType  = 0 // Requirement type token.
	argValue = 1 // Requirement value.
)

// Number of positional arguments a kernel grant requires.
const kernelArgCount = 2
