package manifest

import (
	"strconv"

	"github.com/cruciblehq/crux/crex"
)

// A build stage in a recipe.
//
// Each stage is an independent build pipeline with its own source image and
// steps. Named stages can be referenced from subsequent stages (e.g., a step
// name "builder" can be referenced as a source in later stages with a path
// like "builder:/app/bin"). The last stage in a recipe is the output stage;
// its image is exported as the final build artifact. All preceding stages are
// transient and exist only to produce intermediary artifacts. When Platform
// is set, the stage only runs for the matching target platform. Steps within
// a platform-scoped stage cannot use [Step.Platform] since the platform is
// already fixed for the entire stage.
type Stage struct {

	// Identifies the stage for cross-stage references.
	//
	// When set, must be unique across all stages in the recipe. Used as the
	// prefix in copy source paths (e.g. "builder:/path"). Stages that do not
	// need to be referenced by other stages can omit the name.
	Name string `json:"name,omitempty"`

	// Restricts this stage to a specific target platform.
	//
	// When set, the stage is only built when the target platform matches. The
	// format is "os/arch" (e.g. "linux/arm64"). Steps within a platform-scoped
	// stage cannot use [Step.Platform].
	Platform string `json:"platform,omitempty"`

	// Base image for this stage.
	//
	// A Crucible resource reference (e.g. "crucible/runtime 0.1.0"). When nil,
	// the stage starts from an empty filesystem (scratch).
	From *Ref `json:"from,omitempty"`

	// Security grants for this stage.
	//
	// Domain grants (starting with ".") target a specific subsystem directly.
	// Reference grants (bare name, no ".") name an affordance resource whose
	// grants are inlined at build time.
	Grants []Grant `json:"grants,omitempty"`

	// Ordered build steps for this stage.
	//
	// Each step is executed sequentially in the build container. Steps can
	// reference artifacts from previous stages using the stage name as a
	// prefix in copy sources (e.g. "builder:/app/bin"). When the stage has
	// a platform selector, steps cannot use [Step.Platform].
	Steps []Step `json:"steps,omitempty"`
}

// Validates the stage.
//
// When [Stage.From] is set, it must be a valid ref. A nil From indicates a
// scratch stage with an empty filesystem. Each step is validated recursively
// with positional context. When [Stage.Platform] is set, steps cannot use
// [Step.Platform]. Grants are also validated recursively.
func (s *Stage) Validate() error {
	if s.Name != "" {
		if _, err := strconv.Atoi(s.Name); err == nil {
			return crex.Wrap(ErrInvalidStage, ErrNumericStageName)
		}
	}

	if s.From != nil {
		if err := s.From.Validate(); err != nil {
			return crex.Wrap(ErrInvalidStage, err)
		}
	}

	for i := range s.Steps {
		if s.Platform != "" && stepUsesPlatform(&s.Steps[i]) {
			return crex.Wrapf(ErrInvalidStage, "step %d: %w", i+1, ErrPlatformInPlatformStage)
		}
		if err := s.Steps[i].Validate(); err != nil {
			return crex.Wrapf(ErrInvalidStage, "step %d: %w", i+1, err)
		}
	}

	for i := range s.Grants {
		if err := s.Grants[i].Validate(); err != nil {
			return crex.Wrapf(ErrInvalidStage, "grant %d: %w", i+1, err)
		}
	}

	return nil
}

// Whether a step or any of its children use the platform field.
func stepUsesPlatform(s *Step) bool {
	if s.Platform != "" {
		return true
	}
	for i := range s.Steps {
		if stepUsesPlatform(&s.Steps[i]) {
			return true
		}
	}
	return false
}
