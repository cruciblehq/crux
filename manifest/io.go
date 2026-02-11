package manifest

import (
	"os"

	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/resource"
	"github.com/go-viper/mapstructure/v2"
	"gopkg.in/yaml.v3"
)

// Loads and parses a manifest file.
//
// The path parameter specifies the full path to the YAML manifest file. The
// function reads and unmarshals the file contents according to the [Manifest]
// structure. Returns the parsed [Manifest] on success, or an error if the file
// could not be read or parsed.
func Read(path string) (*Manifest, error) {

	// Read file into raw map
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, crex.Wrap(ErrManifestReadFailed, err)
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, crex.Wrap(ErrManifestReadFailed, err)
	}

	// Decode into Manifest struct
	var m Manifest
	if err := decodeManifest(raw, &m); err != nil {
		return nil, crex.Wrap(ErrManifestReadFailed, err)
	}

	return &m, nil
}

// Decodes a raw map into a [Manifest] structure.
//
// The raw parameter is a map representing the unmarshaled content. The manifest
// parameter is a pointer to the [Manifest] structure where the decoded data
// should be stored. The function first decodes common fields into the manifest,
// then resolves the resource type to determine the concrete manifest type.
func decodeManifest(raw map[string]any, manifest *Manifest) error {

	// Decode common fields
	if err := decodeMap(raw, manifest); err != nil {
		return err
	}

	// Resolve type-specific config
	configs := map[resource.Type]any{
		resource.TypeRuntime: &Runtime{},
		resource.TypeService: &Service{},
		resource.TypeWidget:  &Widget{},
	}

	target, ok := configs[manifest.Resource.Type]
	if !ok {
		return ErrUnknownResourceType
	}

	// Decode type-specific config
	if err := decodeMap(raw, target); err != nil {
		return err
	}

	// Validate type-specific config
	if v, ok := target.(validator); ok {
		if err := v.validate(); err != nil {
			return err
		}
	}

	// Assign to manifest
	manifest.Config = target

	return nil
}

// Decodes a raw map into a target struct using yaml struct tags.
func decodeMap(raw map[string]any, target any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  target,
		TagName: "yaml",
		Squash:  true,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(raw)
}
