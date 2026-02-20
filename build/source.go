package build

import (
	"strings"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/reference"
)

// Source type for a stage's base image.
type sourceType string

const (

	// Local OCI image archive.
	sourceFile sourceType = "file"

	// Crucible runtime reference.
	sourceRef sourceType = "ref"
)

// Describes the origin of a stage's base image.
//
// Type discriminates between a local OCI tarball and a Crucible runtime
// reference. Value holds the raw payload after the type prefix.
type source struct {

	// Discriminates between file and ref sources.
	Type sourceType

	// Source payload after the type prefix.
	//
	// For file sources this is the local OCI image archive path relative to
	// the manifest file. For ref sources this is the Crucible runtime
	// reference string as passed to [reference.Parse].
	Value string
}

// Parses a [manifest.Stage.From] string into a [source].
//
// The string is tokenized on whitespace, so tabs and multiple spaces are
// treated identically to a single space. A "file" prefix selects a local OCI
// archive. Everything else is parsed as a Crucible runtime reference via
// [reference.Parse], with the optional "ref" prefix stripped first. A runtime
// literally named "file" must use the "ref" prefix to avoid ambiguity.
func parseFrom(from string, options reference.IdentifierOptions) (source, error) {
	fields := strings.Fields(from)
	if len(fields) == 0 {
		return source{}, ErrInvalidFromFormat
	}

	switch fields[0] {
	case "file":
		if len(fields) < 2 {
			return source{}, ErrInvalidFromFormat
		}
		path := strings.Join(fields[1:], " ")
		return source{Type: sourceFile, Value: path}, nil

	case "ref":
		if len(fields) < 3 {
			return source{}, ErrInvalidFromFormat
		}
		value := strings.Join(fields[1:], " ")
		if _, err := reference.Parse(value, string(manifest.TypeRuntime), options); err != nil {
			return source{}, crex.Wrap(ErrInvalidFromFormat, err)
		}
		return source{Type: sourceRef, Value: value}, nil

	default:
		value := strings.Join(fields, " ")
		if _, err := reference.Parse(value, string(manifest.TypeRuntime), options); err != nil {
			return source{}, crex.Wrap(ErrInvalidFromFormat, err)
		}
		return source{Type: sourceRef, Value: value}, nil
	}
}
