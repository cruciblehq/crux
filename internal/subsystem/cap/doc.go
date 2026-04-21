// Package cap implements the Linux capabilities subsystem for ward.
//
// Capability policies declare which Linux capabilities a container process
// is granted at runtime. Each grant names one capability from the kernel's
// capability set. The package provides a catalog of known capabilities and
// functions to build grants from capability rules. The rules are strings
// that specify the capability name without the CAP_ prefix. The package also
// defines the subsystem-specific rule expression for capability grants and
// functions to encode and decode these rules into a binary wire format for
// communication with the ward daemon.
package cap
