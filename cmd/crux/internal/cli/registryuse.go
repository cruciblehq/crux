package cli

import (
	"context"
	"log/slog"
)

// Represents the 'crux registry use' command.
type RegistryUseCmd struct {
	Name string `arg:"" help:"Name of the registry to use by default."`
}

// Makes a registry the default for commands that take --registry.
func (c *RegistryUseCmd) Run(_ context.Context) error {
	cfg, err := loadRegistryConfig()
	if err != nil {
		return err
	}
	if _, ok := cfg.find(c.Name); !ok {
		return unknownRegistryError(c.Name)
	}

	cfg.Default = c.Name
	if c.Name == builtinRegistry {
		cfg.Default = ""
	}
	if err := cfg.save(); err != nil {
		return err
	}

	slog.Info("default registry set", "name", c.Name)
	return nil
}
