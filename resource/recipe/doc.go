// Package recipe provides the interface and orchestration logic for building
// OCI images from Crucible recipe stages.
//
// The [Backend] interface is the compute-backend contract. It is implemented
// by the compute package using containerd.
//
// The [Builder] drives the stage pipeline: it imports base images, compiles
// stage affordances into a security policy, and executes each step in order.
//
//	 b := recipe.NewBuilder(source, workdir, backend)
//	buildDir, err := b.Run(ctx, manifest, &manifest.Recipe, output, entrypoint)
package recipe
