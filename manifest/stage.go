package manifest

import (
	"github.com/cruciblehq/crux/codec"
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
	Name string `codec:"name,omitempty"`

	// Restricts this stage to a specific target platform.
	//
	// When set, the stage is only built when the target platform matches. The
	// format is "os/arch" (e.g. "linux/arm64"). Steps within a platform-scoped
	// stage cannot use [Step.Platform].
	Platform string `codec:"platform,omitempty"`

	// Base image for this stage.
	//
	// A Crucible resource reference string (e.g. "crucible/runtime 0.1.0").
	// When empty, the stage starts from an empty filesystem (scratch).
	From string `codec:"from,omitempty"`

	// Arguments for the base image reference.
	//
	// Maps parameter names to string values for the base image affordance. Must
	// not be set when [Stage.From] is empty. Each key must be a valid name.
	Args Args `codec:"args,omitempty"`

	// Grants for this stage.
	//
	// Domain grants (starting with ".") target a specific subsystem directly.
	// Reference grants (bare name, no ".") name an affordance resource whose
	// grants are inlined at build time. Grant scopes allow platform-specific
	// grants within a universal stage. Platform-scoped stages cannot contain
	// grant scopes with a platform selector.
	Grants []GrantScope `codec:"grants,omitempty"`

	// Ordered build steps for this stage.
	//
	// Each step is executed sequentially in the build container. Steps can
	// reference artifacts from previous stages using the stage name as a
	// prefix in copy sources (e.g. "builder:/app/bin"). When the stage has
	// a platform selector, steps cannot use [Step.Platform].
	Steps []Step `codec:"steps,omitempty"`
}

// Validates the stage.
//
// When [Stage.From] is set, arg keys must be valid names. Args must not be set
// when From is empty. An empty From indicates a scratch stage. Each step is
// validated recursively with positional context. When [Stage.Platform] is set,
// steps cannot use [Step.Platform] and grant scopes cannot carry a platform
// selector. Grants are also validated recursively.
func (s *Stage) Validate() error {
	if s.Name != "" && !isValidName(s.Name) {
		return crex.Wrap(ErrInvalidStage, ErrInvalidStageName)
	}

	if s.Platform != "" && !isValidPlatform(s.Platform) {
		return crex.Wrap(ErrInvalidStage, ErrInvalidPlatform)
	}

	if len(s.Args) > 0 && s.From == "" {
		return crex.Wrap(ErrInvalidStage, ErrMissingFrom)
	}
	if err := s.Args.Validate(); err != nil {
		return crex.Wrap(ErrInvalidStage, err)
	}

	for i := range s.Steps {
		if err := s.validateStep(i); err != nil {
			return err
		}
	}

	for i := range s.Grants {
		if err := s.validateGrantScope(i); err != nil {
			return err
		}
	}

	return nil
}

// Validates a single step by index.
//
// When the stage has a platform selector, steps cannot use [Step.Platform].
// Validation errors are wrapped with the step index for context.
func (s *Stage) validateStep(i int) error {
	if s.Platform != "" && stepUsesPlatform(&s.Steps[i]) {
		return crex.Wrapf(ErrInvalidStage, "step %d: %w", i+1, ErrPlatformInPlatformStage)
	}
	if err := s.Steps[i].Validate(); err != nil {
		return crex.Wrapf(ErrInvalidStage, "step %d: %w", i+1, err)
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

// Validates a single grant scope by index.
//
// When the stage has a platform selector, grant scopes cannot carry a
// platform selector. Validation errors are wrapped with the grant scope
// index for context.
func (s *Stage) validateGrantScope(i int) error {
	if s.Platform != "" && s.Grants[i].Platform != "" {
		return crex.Wrapf(ErrInvalidStage, "grant scope %d: %w", i+1, ErrGrantScopePlatformInPlatformStage)
	}
	if err := s.Grants[i].Validate(); err != nil {
		return crex.Wrapf(ErrInvalidStage, "grant scope %d: %w", i+1, err)
	}
	return nil
}

// Encodes the stage to a format-independent value.
//
// Grants are serialized in the same flat format used by [Affordance.Scopes]:
// universal grants as plain strings, platform-scoped groups as objects with
// a platform key and a nested grants list.
func (s *Stage) Encode() (any, error) {
	m := make(map[string]any)
	if s.Name != "" {
		m["name"] = s.Name
	}
	if s.Platform != "" {
		m["platform"] = s.Platform
	}
	if s.From != "" {
		m["from"] = s.From
	}
	if len(s.Args) > 0 {
		m["args"] = s.Args
	}
	if len(s.Grants) > 0 {
		list, err := encodeScopes(s.Grants)
		if err != nil {
			return nil, err
		}
		m["grants"] = list
	}
	if len(s.Steps) > 0 {
		steps := make([]any, len(s.Steps))
		for i := range s.Steps {
			sm, err := codec.ToMap(&s.Steps[i])
			if err != nil {
				return nil, err
			}
			steps[i] = sm
		}
		m["steps"] = steps
	}
	return m, nil
}

// Decodes the stage from a raw parsed map.
//
// Grants are decoded from the flat format: plain strings for universal grants
// and objects with a platform key for platform-scoped groups.
func (s *Stage) Decode(raw any) error {
	src, ok := raw.(map[string]any)
	if !ok {
		return crex.Wrapf(ErrInvalidStage, "unexpected type %T", raw)
	}
	if err := codec.Field(src, s, "Name"); err != nil {
		return crex.Wrap(ErrInvalidStage, err)
	}
	if err := codec.Field(src, s, "Platform"); err != nil {
		return crex.Wrap(ErrInvalidStage, err)
	}
	if err := codec.Field(src, s, "From"); err != nil {
		return crex.Wrap(ErrInvalidStage, err)
	}
	if err := codec.Field(src, s, "Args"); err != nil {
		return crex.Wrap(ErrInvalidStage, err)
	}
	if err := codec.Field(src, s, "Steps"); err != nil {
		return crex.Wrap(ErrInvalidStage, err)
	}
	list, _ := src["grants"].([]any)
	if len(list) > 0 {
		scopes, err := decodeScopes(list)
		if err != nil {
			return crex.Wrap(ErrInvalidStage, err)
		}
		s.Grants = scopes
	}
	return nil
}
