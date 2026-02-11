package manifest

import (
	"fmt"
	"strings"

	"github.com/cruciblehq/crux/reference"
	"github.com/cruciblehq/crux/resource"
)

// Describes a base image and the build steps applied on top of it.
//
// A recipe is the reusable unit shared by resource types that produce OCI
// images. It pairs a source image ([Recipe.From]) with an ordered list of
// build steps ([Recipe.Steps]).
type Recipe struct {

	// Specifies the base image source.
	//
	// The value is a local OCI tarball prefixed with "file", or a Crucible
	// runtime reference optionally prefixed with "ref".
	From string `yaml:"from"`

	// Ordered build steps applied on top of the base image.
	//
	// Steps are executed sequentially in the order they appear. See [Step]
	// for the supported fields.
	Steps []Step `yaml:"steps"`
}

// Validates the recipe.
//
// Checks that [Recipe.From] is present and that every step has valid field
// combinations. Called automatically during [Read] after decoding via the
// embedding type's validate method.
func (r *Recipe) validate() error {
	if strings.TrimSpace(r.From) == "" {
		return ErrMissingFrom
	}

	for i, s := range r.Steps {
		if err := s.validate(); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}
	return nil
}

// Parses the [Recipe.From] string into a [RuntimeSource].
//
// The string is tokenized on whitespace, so tabs and multiple spaces are
// treated identically to a single space. A "file" prefix selects a local OCI
// archive. Everything else is parsed as a Crucible runtime reference via
// [reference.Parse], with the optional "ref" prefix stripped first. A runtime
// literally named "file" must use the "ref" prefix to avoid ambiguity.
func (r *Recipe) ParseFrom(options reference.IdentifierOptions) (RuntimeSource, error) {
	fields := strings.Fields(r.From)
	if len(fields) == 0 {
		return RuntimeSource{}, ErrInvalidFromFormat
	}

	switch fields[0] {
	case "file":
		if len(fields) < 2 {
			return RuntimeSource{}, ErrInvalidFromFormat
		}
		path := strings.Join(fields[1:], " ")
		return RuntimeSource{Type: RuntimeSourceFile, Value: path}, nil

	case "ref":
		if len(fields) < 3 {
			return RuntimeSource{}, ErrInvalidFromFormat
		}
		value := strings.Join(fields[1:], " ")
		ref, err := reference.Parse(value, resource.TypeRuntime, options)
		if err != nil {
			return RuntimeSource{}, fmt.Errorf("%w: %w", ErrInvalidFromFormat, err)
		}
		return RuntimeSource{Type: RuntimeSourceRef, Value: value, Ref: ref}, nil

	default:
		value := strings.Join(fields, " ")
		ref, err := reference.Parse(value, resource.TypeRuntime, options)
		if err != nil {
			return RuntimeSource{}, fmt.Errorf("%w: %w", ErrInvalidFromFormat, err)
		}
		return RuntimeSource{Type: RuntimeSourceRef, Value: value, Ref: ref}, nil
	}
}
