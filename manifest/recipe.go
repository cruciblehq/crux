package manifest

import (
	"fmt"
	"strings"

	"github.com/cruciblehq/crux/kit/crex"
)

// Describes a build pipeline as one or more stages.
//
// A recipe is the reusable unit shared by resource types that produce OCI
// images. It contains a list of [Recipe.Stages], each with its own source
// image and build steps. Exactly one stage must be non-transient, which is
// the stage that is exported as the final build artifact.
type Recipe struct {

	// Build stages.
	//
	// Each stage is an independent build pipeline with its own source image
	// and steps. Stages run in declaration order. Artifacts produced by a
	// stage can be referenced from subsequent stages via the stage name in
	// a copy source (e.g. "builder:/app/bin").
	Stages []Stage `yaml:"stages"`
}

// A build stage in a recipe.
//
// Each stage is an independent build pipeline with its own source image
// and steps. Named stages can be referenced from subsequent stages
// (e.g. "builder:/app/bin" in a copy step). Stages are non-transient by
// default, meaning their image is exported as the final build artifact.
// Set [Stage.Transient] to true for intermediate stages.
type Stage struct {

	// Identifies the stage for cross-stage references.
	//
	// When set, must be unique across all stages in the recipe. Used as
	// the prefix in copy source paths (e.g. "builder:/path"). Stages that
	// do not need to be referenced by other stages can omit the name.
	Name string `yaml:"name,omitempty"`

	// Marks this stage as an intermediate build helper.
	//
	// Transient stages are not exported as the final build artifact. They
	// exist only to produce artifacts that are copied into later stages.
	// In a single-stage recipe this field can be omitted (defaults to
	// false). In a multi-stage recipe every stage except the output stage
	// must be marked transient.
	Transient bool `yaml:"transient,omitempty"`

	// Specifies the base image source for this stage.
	From string `yaml:"from"`

	// Ordered build steps for this stage.
	Steps []Step `yaml:"steps"`
}

// Validates the recipe.
//
// Checks that at least one stage exists, that named stages have unique names,
// and that exactly one stage is non-transient (the output stage).
func (r *Recipe) validate() error {
	if len(r.Stages) == 0 {
		return ErrMissingStages
	}

	seen := make(map[string]bool, len(r.Stages))
	outputStages := 0

	for i, stage := range r.Stages {
		if err := validateStage(stage, i, seen); err != nil {
			return err
		}
		if !stage.Transient {
			outputStages++
		}
	}

	if outputStages == 0 {
		return ErrNoOutputStage
	}
	if outputStages > 1 {
		return ErrMultipleOutputStages
	}

	return nil
}

// Validates a single stage's name, from, and steps.
func validateStage(stage Stage, index int, seen map[string]bool) error {
	label := stageLabel(stage.Name, index)

	if name := strings.TrimSpace(stage.Name); name != "" {
		if seen[name] {
			return crex.Wrap(fmt.Errorf("stage %s", label), ErrDuplicateStageName)
		}
		seen[name] = true
	}

	if strings.TrimSpace(stage.From) == "" {
		return crex.Wrap(fmt.Errorf("stage %s", label), ErrMissingFrom)
	}

	for j, s := range stage.Steps {
		if err := s.validate(); err != nil {
			return crex.Wrap(fmt.Errorf("stage %s step %d", label, j+1), err)
		}
	}

	return nil
}

// Returns a label for a stage, preferring the name when available and falling
// back to the 1-based index.
func stageLabel(name string, index int) string {
	if name != "" {
		return fmt.Sprintf("%q", name)
	}
	return fmt.Sprintf("%d", index+1)
}