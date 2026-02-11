package manifest

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cruciblehq/crux/oci"
)

// A build step in a runtime definition.
//
// Fields are either operations or modifiers. Operations are the actions:
// [Step.Run], [Step.Exec], [Step.Copy]; they are mutually exclusive.
// Modifiers are [Step.Shell], [Step.Env], [Step.Workdir], [Step.Platform];
// paired with an operation they apply to that single step, standalone they
// persist in the image for subsequent steps. Modifiers combine freely with
// each other. [Step.Platform] with [Step.Steps] creates a group whose
// children inherit any modifiers set at the group level. Invalid combinations
// are rejected during validation.
type Step struct {

	// Executes a command through a shell inside the build container.
	//
	// The command string is passed to the default shell (/bin/sh) or to
	// the shell specified by [Step.Shell]. It runs as root in the current
	// working directory.
	Run string `yaml:"run,omitempty"`

	// Selects the shell used to execute [Step.Run] commands.
	//
	// When paired with [Step.Run], overrides the default shell for that
	// single command. When set alone or with other modifiers, changes the
	// default shell for all subsequent run operations. Defaults to /bin/sh.
	Shell string `yaml:"shell,omitempty"`

	// Executes a binary directly without a shell.
	//
	// The first element is the binary path, the rest are arguments. No shell
	// interpretation is performed.
	Exec []string `yaml:"exec,omitempty"`

	// Copies a file or directory from the host into the image.
	//
	// Specified as "src dest" where src is a host path relative to the
	// manifest file and dest is a path inside the image. A relative dest
	// resolves against the current working directory. Directories are
	// copied recursively.
	Copy string `yaml:"copy,omitempty"`

	// Sets environment variables in the build container.
	//
	// When paired with an operation, the variables are scoped to that single
	// command and not persisted. When set alone, the variables persist in the
	// image and are inherited by subsequent steps and any service that uses
	// this runtime as its base.
	Env map[string]string `yaml:"env,omitempty"`

	// Sets the working directory inside the build container.
	//
	// When paired with an operation, overrides the working directory for that
	// single step without changing the default. When set alone, changes the
	// default working directory for all subsequent steps and persists it in
	// the image configuration.
	Workdir string `yaml:"workdir,omitempty"`

	// Restricts this step or group to a specific platform.
	//
	// When set with an operation or modifier, restricts it to the given
	// platform. When set with [Step.Steps], creates a platform-scoped
	// group; other modifiers on the same step apply to all children in
	// the group. The format is "os/arch" (e.g. "linux/amd64").
	Platform string `yaml:"platform,omitempty"`

	// Child steps scoped to the platform specified by [Step.Platform].
	//
	// When set, [Step.Platform] must also be set. Children inherit any
	// modifiers set at the group level and follow the same rules as
	// top-level steps.
	Steps []Step `yaml:"steps,omitempty"`
}

// Validates the step and its children.
//
// Delegates to [Step.validateStructure], [Step.validateModifiers], and
// [Step.validateValues] to check field combinations, then recursively
// validates child steps. Children cannot set [Step.Platform].
func (s *Step) validate() error {
	if err := s.validateStructure(); err != nil {
		return err
	}
	if err := s.validateModifiers(); err != nil {
		return err
	}
	if err := s.validateValues(); err != nil {
		return err
	}

	for i, child := range s.Steps {
		if child.Platform != "" {
			return fmt.Errorf("step %d: %w", i+1, ErrNestedPlatform)
		}
		if err := child.validate(); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}

	return nil
}

// Validates structural rules for the step.
//
// Ensures at least one field is set, that operations are mutually exclusive,
// that [Step.Steps] is paired with [Step.Platform], and that operations do
// not carry child steps.
func (s *Step) validateStructure() error {
	hasRun := s.Run != ""
	hasExec := len(s.Exec) > 0
	hasCopy := s.Copy != ""
	hasMod := s.Shell != "" || len(s.Env) > 0 || s.Workdir != "" || s.Platform != ""
	hasSteps := len(s.Steps) > 0

	if !hasRun && !hasExec && !hasCopy && !hasMod && !hasSteps {
		return ErrEmptyStep
	}
	if (hasRun && hasExec) || (hasRun && hasCopy) || (hasExec && hasCopy) {
		return ErrMultipleOperations
	}
	if hasSteps && s.Platform == "" {
		return ErrStepsWithoutPlatform
	}
	if (hasRun || hasExec || hasCopy) && hasSteps {
		return ErrOperationWithSteps
	}
	return nil
}

// Validates modifier compatibility with the current operation.
//
// Rejects [Step.Shell] when paired with [Step.Exec] or [Step.Copy], and
// [Step.Env] when paired with [Step.Copy].
func (s *Step) validateModifiers() error {
	hasCopy := s.Copy != ""
	hasShell := s.Shell != ""
	hasEnv := len(s.Env) > 0

	if hasShell && len(s.Exec) > 0 {
		return ErrShellWithExec
	}
	if hasShell && hasCopy {
		return ErrShellWithCopy
	}
	if hasEnv && hasCopy {
		return ErrEnvWithCopy
	}
	return nil
}

// Validates individual field values.
//
// Checks that [Step.Exec] has a non-empty first element, that [Step.Copy] is
// in valid "src dest" format, and that [Step.Platform] is a known platform.
func (s *Step) validateValues() error {
	if len(s.Exec) > 0 && s.Exec[0] == "" {
		return ErrExecEmpty
	}
	if s.Copy != "" {
		if _, _, err := s.parseCopy(); err != nil {
			return err
		}
	}
	if s.Platform != "" {
		return validatePlatform(s.Platform)
	}
	return nil
}

// Validates a platform string against the known set.
//
// Checks that the string is in valid "os/arch" format and that it matches one
// of the platforms returned by [oci.RequiredPlatforms].
func validatePlatform(platform string) error {
	if _, _, err := oci.ParsePlatform(platform); err != nil {
		return err
	}
	if !slices.Contains(oci.RequiredPlatforms(), platform) {
		return fmt.Errorf("%w: %s", ErrUnknownPlatform, platform)
	}
	return nil
}

// Parses the [Step.Copy] string into source and destination paths.
//
// The input must be in "src dest" format. Returns an error if the format is
// invalid. The source path is relative to the manifest file. The destination
// path is either absolute or relative to the working directory.
func (s *Step) parseCopy() (src, dest string, err error) {
	parts := strings.Fields(s.Copy)
	if len(parts) != 2 {
		return "", "", ErrInvalidCopyFormat
	}
	return parts[0], parts[1], nil
}
