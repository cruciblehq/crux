// Package kernel implements the .kernel AGL subsystem.
//
// A .kernel grant declares a requirement on the VM's Linux kernel, allowing
// services to specify the kernel features and configurations they require.
//
// Grant syntax:
//
//	.kernel config  FEATURE  — CONFIG_* flag (without CONFIG_ prefix)
//	.kernel module  NAME     — kernel module that must be available
//	.kernel version VER      — minimum kernel version (e.g. 5.15)
//	.kernel boot    PARAM    — boot parameter that must appear in /proc/cmdline
//	.kernel lsm     NAME     — Linux Security Module that must be active
//	.kernel hw      FLAG     — CPU/hardware feature flag (from /proc/cpuinfo)
package kernel
