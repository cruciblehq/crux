// Package cap implements the Linux capabilities subsystem.
//
// Capability grants declare which Linux capabilities a container process
// receives at startup. Each grant names one capability from the kernel's
// capability set, written without the CAP_ prefix, and an optional mode that
// selects which of the five per-task capability sets (effective, permitted,
// inheritable, bounding, ambient) are populated. The catalog of valid
// capability names lives in the shared package and is reused by both cap
// and fcap so that the two subsystems agree on what the kernel exposes.
//
// A [Subsystem] accumulates grants and produces the resulting OCI runtime-spec
// capability set, while [Compose] merges that set into the capability sub-tree
// of an existing spec, deduplicating entries within each of the five sets.
//
//	b := cap.New(nil)
//	if err := b.Build(g); err != nil {
//		log.Fatal(err)
//	}
//	spec := b.Spec()
//	b.Compose(spec.Process.Capabilities)
package cap
