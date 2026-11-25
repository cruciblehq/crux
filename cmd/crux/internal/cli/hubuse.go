package cli

import (
	"context"
	"log/slog"
)

// Represents the 'crux hub use' command.
type HubUseCmd struct {
	Name string `arg:"" help:"Name of the hub to use by default."`
}

// Makes a hub the default for commands that take --hub.
func (c *HubUseCmd) Run(_ context.Context) error {
	cfg, err := loadHubConfig()
	if err != nil {
		return err
	}
	if _, ok := cfg.find(c.Name); !ok {
		return unknownHubError(c.Name)
	}

	cfg.Default = c.Name
	if c.Name == builtinHub {
		cfg.Default = ""
	}
	if err := cfg.save(); err != nil {
		return err
	}

	slog.Info("default hub set", "name", c.Name)
	return nil
}
