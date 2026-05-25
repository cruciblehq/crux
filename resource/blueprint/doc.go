// Package blueprint implements the deployment plan compilation pipeline for
// Crucible blueprints.
//
// [Build] resolves service references, collects affordances, and writes
// plan.yaml to the output directory.
//
//	err := blueprint.Build(ctx, cfg, "production", source, "build")
package blueprint
