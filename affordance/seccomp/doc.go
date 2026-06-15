// Package seccomp implements the seccomp syscall filter subsystem.
//
// Seccomp grants declare which syscalls the container process may invoke.
// Each grant names a syscall and may further restrict it with a curated
// sub-filter for multi-purpose syscalls such as ioctl, fcntl, and prctl,
// for example "ioctl tty", "fcntl flags", or "prctl name". Sub-filters
// resolve through a catalog that maps friendly names to the safe argument
// values understood by the kernel, so policy authors do not have to
// memorise raw constant numbers.
//
// A [Subsystem] wraps the OCI seccomp section of the unified spec and
// accumulates allow rules in place. Build appends entries for one parsed
// grant; both operations dedupe
// structurally identical rules and collapse filtered rules that are
// subsumed by an unconditional allow for the same syscall.
//
//	s := seccomp.New(spec.Linux.Seccomp)
//	if err := s.Build(g); err != nil {
//		return err
//	}
package seccomp
