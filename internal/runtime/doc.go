// Package runtime dispatches grants to the appropriate subsystem.
//
// Each subsystem (cap, rlimit, seccomp, fcap, cgroup, mac) is a separate
// implementation that compiles parsed grants into a slice of the unified
// [shared.Spec] model. The aggregate [Builder] owns one [shared.Spec] per
// session and one subsystem implementation per slice, and routes incoming
// grants to the correct subsystem based on the subsystem name carried by
// each grant. After all grants have been ingested, [Builder.Spec] returns
// the aggregated model.
//
// Building an aggregate spec from a sequence of grants:
//
//	b := runtime.NewBuilder()
//	for _, g := range grants {
//		if err := b.Build(&g); err != nil {
//			log.Fatal(err)
//		}
//	}
//	spec := b.Spec()
package runtime
