package blueprint

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Loads a blueprint from a YAML file.
//
// The path parameter specifies the full path to the blueprint file.
func Read(path string) (*Blueprint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var bp Blueprint
	if err := yaml.Unmarshal(data, &bp); err != nil {
		return nil, err
	}
	return &bp, nil
}

// Saves the blueprint to a YAML file.
func (bp *Blueprint) Write(path string) error {
	data, err := yaml.Marshal(bp)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
