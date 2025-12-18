package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

const (

	// The path of the manifest file within a Crucible resource project.
	Manifestfile = ".cruciblerc/manifest.yaml"
)

// Loads and parses a manifest file.
//
// The dir parameter specifies the resource directory where the manifest file
// is located. The function constructs the full path by joining dir with the
// expected manifest file location. It then reads the file, and unmarshals its
// contents according to the [Manifest] structure. The structure is expected to
// conform to the Crucible manifest schema, identified by "field" struct tags.
// Returns the parsed [Manifest] on success, or an error if the file could not
// be read or parsed.
func Read() (*Manifest, error) {

	// Locate the manifest within the current working directory
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Full path to manifest file
	path := filepath.Join(dir, Manifestfile)

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal raw YAML into a map
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	// Decode into Manifest struct
	var m Manifest
	if err := decodeManifest(raw, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

// Decodes a raw map into a Manifest structure.
//
// The raw parameter is a map representing the unmarshaled content. The manifest
// parameter is a pointer to the [Manifest] structure where the decoded data
// should be stored. The function first decodes common fields into the manifest,
// then resolves the resource type to determine the concrete manifest type.
func decodeManifest(raw map[string]any, manifest *Manifest) error {

	// Decode common fields
	if err := decodeAny(raw, manifest); err != nil {
		return err
	}

	// Resolve type-specific config
	configs := map[string]any{
		"widget":  &Widget{},
		"service": &Service{},
	}

	target, ok := configs[manifest.Resource.Type]
	if !ok {
		return fmt.Errorf("%s: %w", manifest.Resource.Type, ErrUnknownResourceType)
	}

	// Decode type-specific config
	if err := decodeAny(raw, target); err != nil {
		return err
	}

	// Assign to manifest
	manifest.Config = target

	return nil
}

// Decodes a raw map into a target structure.
//
// The raw parameter is a map representing the unmarshaled YAML content. The
// target parameter is a pointer to the structure where the decoded data should
// be stored. The function uses "key" struct tags to map fields from the raw
// map to the target structure.
func decodeAny(raw map[string]any, target any) error {

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  target,
		TagName: "key", // <--------
	})

	if err != nil {
		return err
	}

	return decoder.Decode(raw)
}
