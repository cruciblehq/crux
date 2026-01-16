package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/pkg/config"
)

// Set the default provider.
type ProviderSetDefaultCmd struct {
	Name string `arg:"" help:"Name of the provider to set as default"` // Name of the provider
}

// Run executes the set-default command.
//
// Sets the specified provider as the default provider. The provider must
// already exist in the configuration, otherwise an error is returned.
func (c *ProviderSetDefaultCmd) Run(ctx context.Context) error {
	cfg, err := config.LoadProviders()
	if err != nil {
		return err
	}

	if err := cfg.SetDefault(c.Name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	slog.Info("Default provider set", "name", c.Name)

	return nil
}
