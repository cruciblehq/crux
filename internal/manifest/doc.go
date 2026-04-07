// Package manifest defines the protocol types for Crucible resource manifests.
//
// A manifest is the top-level document that declares a resource's metadata,
// type, and type-specific configuration. A manifest is represented by the
// [Manifest] struct, which contains a schema version, [Resource] metadata
// (type, name, version), and a type-specific configuration ([Manifest.Config]).
// The config field holds a pointer to one of the concrete types, determined by
// the resource type.
//
// Services and runtimes share a common build pipeline structure called a
// [Recipe]. A recipe consists of one or more [Stage] values, each with its
// own base image and build [Step] values. Stages run in declaration order.
// Artifacts produced by a named stage can be referenced from subsequent stages
// via copy steps (e.g. "builder:/app/bin"). The last stage is the output
// stage, whose image becomes the final build artifact. All preceding stages
// are transient. When building for multiple platforms, the last stage that
// applies to each platform is that platform's output stage.
//
// Build steps within a stage are classified as operations or modifiers:
//
// Operations are the actions that produce side effects in the build container.
// [Step.Run] executes a command through a shell, and [Step.Copy] copies files
// from the host or from another stage into the image. These two are mutually
// exclusive within a single step.
//
// Modifiers adjust the build environment. [Step.Shell] selects the shell for
// run commands, [Step.Env] sets environment variables, [Step.Workdir] sets the
// working directory, [Step.User] sets the process identity, and [Step.Platform]
// restricts the step to a specific OS/architecture. Modifiers combine freely
// with each other. When paired with an operation, they apply to that single
// step. When set alone, they persist in the image for subsequent steps.
//
// Some modifier-operation combinations are invalid. [Step.Shell] and [Step.Env]
// cannot be paired with [Step.Copy], since copy operations do not involve shell
// execution or environment variables.
//
// Setting [Step.Platform] together with [Step.Steps] creates a platform group:
// a set of child steps that all execute under the specified platform. Modifiers
// on the group step apply to all children. A platform group cannot also contain
// an operation, and nesting platform groups is not allowed.
//
// Stages declare a base image through a [Ref], which identifies a Crucible
// resource to use as the starting point for the build. When no base image is
// specified, the stage starts from an empty filesystem (scratch).
//
// Stages also carry affordance refs that describe requirements from the
// runtime. Each affordance ref identifies an affordance resource. Affordances
// are either composed (referencing child affordances) or primitive-backed
// (mapping to a primitive). Arguments are either absent (bare ref), a
// scalar (assigned to the affordance's default parameter), or a mapping of
// parameter names to values. The primitive resolver translates each ref's
// arguments into concrete effects without the developer needing to know the
// implementation details.
//
// Every type in the package exposes a Validate method that checks structural
// correctness. Validation cascades from [Manifest.Validate] down through
// [Resource], the config type, [Recipe], [Stage], [Step], and [Ref].
//
// Encoding a manifest:
//
//	m := &manifest.Manifest{
//		Resource: manifest.Resource{
//			Type:    manifest.TypeService,
//			Name:    "crucible/hub",
//			Version: "1.0.0",
//		},
//		Config: &manifest.Service{ /* ... */ },
//	}
//	data, err := codec.Encode(m, codec.YAML)
//
// Decoding a manifest:
//
//	var m manifest.Manifest
//	if err := codec.Unmarshal(data, &m, codec.YAML); err != nil {
//		log.Fatal(err) // malformed YAML or unknown resource type
//	}
//	if err := m.Validate(); err != nil {
//		log.Fatal(err) // structural validation failure
//	}
//	switch cfg := m.Config.(type) {
//	case *manifest.Service:
//		fmt.Println("service", cfg)
//	case *manifest.Runtime:
//		fmt.Println("runtime", cfg)
//	}
package manifest
