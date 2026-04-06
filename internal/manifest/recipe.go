package manifest

import (
	"fmt"

	"github.com/cruciblehq/crex"
)

// The OCI image artifact produced by recipe-based builds (runtimes and services).
const ImageFile = "image.tar"

// Describes a build pipeline as one or more stages.
//
// A recipe is the reusable unit shared by resource types that produce OCI
// images. It contains a list of stages, each with its own source image and
// build steps. The last stage in the list is the output stage, whose image
// is exported as the final build artifact. All preceding stages are transient.
type Recipe struct {

	// Build stages.
	//
	// Each stage is an independent build pipeline with its own source image and
	// steps. Stages run in declaration order. Artifacts produced by a stage can
	// be referenced from subsequent stages via the stage name in a copy source
	// (e.g. "builder:/app/bin"). The last stage is the output stage.
	Stages []Stage `codec:"stages,omitempty"`
}

// Validates the recipe.
//
// At least one stage is required to produce an output image. Named stages
// must have unique names and each stage must be structurally valid.
func (r *Recipe) Validate() error {
	if len(r.Stages) == 0 {
		return crex.Wrap(ErrInvalidRecipe, ErrMissingOutputStage)
	}

	seen := make(map[string]bool, len(r.Stages))

	for i := range r.Stages {
		stage := &r.Stages[i]
		name := stage.Name

		if name != "" {
			if seen[name] {
				return crex.Wrapf(ErrInvalidRecipe, "%w: %s", ErrDuplicateStageName, name)
			}
			seen[name] = true
		}

		if err := stage.Validate(); err != nil {
			return crex.Wrapf(ErrInvalidRecipe, "stage %s: %w", stageLabel(name, i), err)
		}
	}

	return nil
}

// Returns a label for a stage, preferring the name when available and
// falling back to the 1-based index.
func stageLabel(name string, index int) string {
	if name != "" {
		return fmt.Sprintf("%q", name)
	}
	return fmt.Sprintf("%d", index+1)
}
