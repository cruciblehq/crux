// Package cap implements the Linux capabilities subsystem.
//
// Capability grants declare which Linux capabilities a container receives at
// startup. Each grant names one capability from the kernel's capability set,
// written without the CAP_ prefix, and a mode that selects which of the five
// capability sets are populated (effective, permitted, inheritable, bounding,
// and ambient). The catalog of capability lives in the caps package and is
// reused by both cap and fcap so that the two subsystems agree on what the
// kernel exposes.
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
