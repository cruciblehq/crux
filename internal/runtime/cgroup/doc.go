// Package cgroup implements the cgroup v2 resource limits subsystem.
//
// Cgroup grants declare resource limits enforced by the kernel's cgroup v2
// controllers. Each grant names a controller knob, for example cpu.max or
// memory.high, and supplies the value to write to it. Knob names are
// validated against a catalog of supported controllers so that typos are
// rejected at build time rather than silently ignored by the kernel at
// runtime.
//
// A [Subsystem] accumulates grants into an internal spec and projects them
// onto the resources sub-tree of an OCI runtime spec. Build applies one
// grant at a time, validating the knob, parsing the value, checking it
// against any prior assignment for conflicts, and recording it. Merge
// folds an externally-supplied OCI resources spec into the subsystem by
// re-applying each declared field through the same conflict checker, so
// pre-built specs are subject to the same validation as direct grants.
//
// Defaults are deny-all. Quotas without a grant project as zero on
// commit, so a zero-grant subsystem yields zero CPU time, zero memory,
// zero processes, and no allowed devices. Specs must explicitly grant
// every quota they need.
//
//	s := cgroup.New(resources)
//	if err := s.Build(g); err != nil {
//		log.Fatal(err)
//	}
package cgroup
