package manifest

import (
	"os"

	"github.com/cruciblehq/crux/codec"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/paths"
)

// Asserts the manifest config to T.
func As[T any](m *Manifest) (T, error) {
	cfg, ok := m.Config.(T)
	if !ok {
		var zero T
		return zero, crex.Wrapf(ErrConfigTypeMismatch, "expected manifest type %T", *new(T))
	}
	return cfg, nil
}

// Reads and decodes the manifest at the given path.
func Read(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := codec.Unmarshal(data, &m, codec.YAML); err != nil {
		return nil, err
	}
	return &m, nil
}

// Reads and decodes the manifest from dir/crucible.yaml.
func ReadAt(dir string) (*Manifest, error) {
	return Read(paths.Manifest(dir))
}

// Reads the manifest at path and asserts its config to T.
//
// Returns an error if the file cannot be read, decoded, or if the config
// type does not match T.
func ReadAs[T any](path string) (T, error) {
	m, err := Read(path)
	if err != nil {
		var zero T
		return zero, err
	}
	return As[T](m)
}

// Reads the manifest from dir/crucible.yaml and asserts its config to T.
func ReadAsAt[T any](dir string) (T, error) {
	return ReadAs[T](paths.Manifest(dir))
}

// Encodes a manifest and writes it to the given path.
func Write(m *Manifest, path string) error {
	data, err := codec.Encode(m, codec.YAML)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, paths.DefaultFileMode)
}

// Encodes a manifest and writes it to dir/crucible.yaml.
func WriteAt(m *Manifest, dir string) error {
	return Write(m, paths.Manifest(dir))
}

// Encodes a plan and writes it to the given path.
func WritePlan(p *Plan, path string) error {
	data, err := codec.Encode(p, codec.YAML)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, paths.DefaultFileMode)
}

// Encodes a plan and writes it to dir/plan.yaml.
func WritePlanAt(p *Plan, dir string) error {
	return WritePlan(p, paths.Plan(dir))
}
