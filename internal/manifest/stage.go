package manifest

import (
	"strconv"

	"github.com/cruciblehq/crex"
)

// A build stage in a recipe.
//
// Each stage is an independent build pipeline with its own source image and
// steps. Named stages can be referenced from subsequent stages (e.g., a step
// name "builder" can be referenced as a source in later stages with a path
// like "builder:/app/bin"). The last stage in a recipe is the output stage;
// its image is exported as the final build artifact. All preceding stages are
// transient and exist only to produce artifacts for later stages. When
// [Stage.Platform] is set, the stage only runs for the matching target
// platform. Steps within a platform-scoped stage cannot use [Step.Platform]
// since the platform is already fixed for the entire stage.
type Stage struct {

	// Identifies the stage for cross-stage references.
	//
	// When set, must be unique across all stages in the recipe. Used as the
	// prefix in copy source paths (e.g. "builder:/path"). Stages that do not
	// need to be referenced by other stages can omit the name.
	Name string `codec:"name,omitempty"`

	// Restricts this stage to a specific target platform.
	//
	// When set, the stage is only built when the target platform matches. The
	// format is "os/arch" (e.g. "linux/arm64"). Steps within a platform-scoped
	// stage cannot use [Step.Platform].
	Platform string `codec:"platform,omitempty"`

	// Base image for this stage.
	//
	// A Crucible resource reference (e.g. "crucible/runtime 0.1.0"). When nil,
	// the stage starts from an empty filesystem (scratch).
	From *Ref `codec:"from,omitempty"`

	// Capabilities this stage requires from the platform.
	//
	// The platform resolves each affordance into effects that apply to the
	// container for this stage.
	Affordances []Ref `codec:"affordances,omitempty"`

	// Ordered build steps for this stage.
	Steps []Step `codec:"steps,omitempty"`
}

// Validates the stage.
//
// When [Stage.From] is set, it must be a valid ref. A nil From indicates a
// scratch stage with an empty filesystem. Each step is validated recursively
// with positional context. When [Stage.Platform] is set, steps cannot use
// [Step.Platform]. Affordances are also validated recursively.
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

	for i := range s.Affordances {
		if err := s.Affordances[i].Validate(); err != nil {
			return crex.Wrapf(ErrInvalidStage, "affordance %d: %w", i+1, err)
		}
	}

	return nil
}

// Reports whether a step or any of its children use the platform field.
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
