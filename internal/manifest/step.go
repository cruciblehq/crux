package manifest

import "github.com/cruciblehq/crex"

// A build step in a recipe stage.
//
// Fields are either operations or modifiers. Operations are the actions:
// [Step.Run], [Step.Copy]; they are mutually exclusive. Modifiers are
// [Step.Shell], [Step.Env], [Step.Workdir], [Step.User], [Step.Platform];
// paired with an operation they apply to that single step, standalone they
// persist in the image for subsequent steps. Modifiers combine freely with
// each other. [Step.Platform] with [Step.Steps] creates a group whose children
// inherit any modifiers set at the group level. Invalid combinations are
// rejected during validation.
type Step struct {

	// Executes a command through a shell inside the build container.
	//
	// The command string is passed to the default shell or to the shell
	// specified by [Step.Shell]. It runs in the current working directory.
	Run string `codec:"run,omitempty"`

	// Copies a file or directory from the host into the image.
	//
	// Specified as "src dest" where src is a host path relative to the
	// manifest file and dest is a path inside the image. A relative dest
	// resolves against the current working directory. Directories are
	// copied recursively.
	Copy string `codec:"copy,omitempty"`

	// Selects the shell used to execute [Step.Run] commands.
	//
	// When paired with [Step.Run], overrides the default shell for that
	// single command. When set alone or with other modifiers, changes the
	// default shell for all subsequent run operations. Defaults to /bin/sh.
	Shell string `codec:"shell,omitempty"`

	// Sets environment variables in the build container.
	//
	// When paired with an operation, the variables are scoped to that single
	// command and not persisted. When set alone, the variables persist in the
	// image and are inherited by subsequent steps and any service that uses
	// this runtime as its base.
	Env map[string]string `codec:"env,omitempty"`

	// Sets the working directory inside the build container.
	//
	// When paired with an operation, overrides the working directory for that
	// single step without changing the default. When set alone, changes the
	// default working directory for all subsequent steps and persists it in
	// the image configuration.
	Workdir string `codec:"workdir,omitempty"`

	// Sets the user identity for commands in the build container.
	//
	// When paired with an operation, overrides the user for that single
	// step without changing the default. When set alone, changes the
	// default user for all subsequent steps. The format is "uid:gid"
	// (e.g. "1000:1000"). The UIDs must exist in the container's user
	// namespace mapping.
	User string `codec:"user,omitempty"`

	// Restricts this step or group to a specific platform.
	//
	// When set with an operation or modifier, restricts it to the given
	// platform. When set with [Step.Steps], creates a platform-scoped
	// group; other modifiers on the same step apply to all children in
	// the group. The format is "os/arch" (e.g. "linux/amd64").
	Platform string `codec:"platform,omitempty"`

	// Child steps scoped to the platform specified by [Step.Platform].
	//
	// When set, [Step.Platform] must also be set. Children inherit any
	// modifiers set at the group level and follow the same rules as
	// top-level steps.
	Steps []Step `codec:"steps,omitempty"`
}

// Validates the step.
//
// Checks field combinations for structural validity and modifier compatibility,
// then recursively validates child steps.
func (s *Step) Validate() error {
	if err := s.validateStructure(); err != nil {
		return crex.Wrap(ErrInvalidStep, err)
	}
	if err := s.validateModifiers(); err != nil {
		return crex.Wrap(ErrInvalidStep, err)
	}

	for i := range s.Steps {
		if s.Steps[i].Platform != "" {
			return crex.Wrapf(ErrInvalidStep, "step %d: %w", i+1, ErrNestedPlatformGroup)
		}
		if err := s.Steps[i].Validate(); err != nil {
			return crex.Wrapf(ErrInvalidStep, "step %d: %w", i+1, err)
		}
	}

	return nil
}

// Validates structural rules for the step.
//
// Ensures at least one field is set, that operations are mutually exclusive,
// that child steps are paired with a platform, and that operations do not
// carry child steps.
func (s *Step) validateStructure() error {
	hasRun := s.Run != ""
	hasCopy := s.Copy != ""
	hasMod := s.Shell != "" || len(s.Env) > 0 || s.Workdir != "" || s.User != "" || s.Platform != ""
	hasSteps := len(s.Steps) > 0

	if !hasRun && !hasCopy && !hasMod && !hasSteps {
		return ErrEmptyStep
	}
	if hasRun && hasCopy {
		return ErrMutuallyExclusiveOps
	}
	if hasSteps && s.Platform == "" {
		return ErrStepsWithoutPlatform
	}
	if (hasRun || hasCopy) && hasSteps {
		return ErrPlatformWithOperation
	}
	return nil
}

// Validates modifier compatibility with the current operation.
//
// Rejects shell and env when paired with a copy operation.
func (s *Step) validateModifiers() error {
	if s.Copy == "" {
		return nil
	}

	if s.Shell != "" {
		return ErrShellWithCopy
	}
	if len(s.Env) > 0 {
		return ErrEnvWithCopy
	}
	return nil
}
