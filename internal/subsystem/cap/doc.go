// Package cap implements the Linux capabilities subsystem.
//
// Capability policies declare which Linux capabilities a container process
// is granted at runtime. Each grant names one capability from the kernel's
// capability set. The package provides a catalog of known capabilities and
// functions to build grants from capability rules. The rules are strings
// that specify the capability name without the CAP_ prefix. The package also
// defines the subsystem-specific rule expression for capability grants and
// functions to parse and compile these rules into capability state.
package cap
