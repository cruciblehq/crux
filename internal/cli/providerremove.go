package cli

import (
	"context"
	"log/slog"

	"github.com/cruciblehq/crux/pkg/config"
)

// Remove a provider configuration.
type ProviderRemoveCmd struct {
	Name string `arg:"" help:"Name of the provider to remove"`
}

// Run executes the provider remove command.
//
// Removes the specified provider configuration. If the removed provider was
// the default provider, the default is cleared and a warning is displayed.
func (c *ProviderRemoveCmd) Run(ctx context.Context) error {
	cfg, err := config.LoadProviders()
	if err != nil {
		return err
	}

	wasDefault := cfg.Default == c.Name

	if err := cfg.RemoveProvider(c.Name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	slog.Info("Provider removed successfully", "name", c.Name)

	if wasDefault {
		slog.Warn("The removed provider was the default provider, so no default provider is set now. Use 'crux provider set-default <name>' to set a new default.", "name", c.Name)
	}

	return nil
}
