package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cruciblehq/crux/config"
)

// Represents the 'crux provider list' command.
type ProviderListCmd struct{}

// Executes the provider list command.
//
// Loads and displays all configured cloud provider details, including type,
// region (for AWS), authentication method, and whether they are set as the
// default provider.
func (c *ProviderListCmd) Run(ctx context.Context) error {
	slog.Info("listing configured providers...")

	cfg, err := config.LoadProviders()
	if err != nil {
		return err
	}

	providers := cfg.ListProviders()
	if len(providers) == 0 {
		slog.Warn("No providers configured. Run 'crux provider add <name>' to configure one.")
		return nil
	}

	changeLine := false

	for _, provider := range providers {

		if changeLine {
			fmt.Println()
		}

		isDefault := cfg.Default == provider.Name
		defaultMarker := ""
		if isDefault {
			defaultMarker = " (default)"
		}

		fmt.Printf("Name: %s%s\n", provider.Name, defaultMarker)
		fmt.Printf("Type: %s\n", provider.Type)

		switch cfg := provider.Config.(type) {
		case *config.AWSProvider:
			if cfg.Region != "" {
				fmt.Printf("Region: %s\n", cfg.Region)
			}
			fmt.Printf("Auth Method: %s\n", cfg.AuthMethod)
			switch auth := cfg.Auth.(type) {
			case *config.AWSProfileAuth:
				fmt.Printf("Profile: %s\n", auth.Profile)
			case *config.AWSKeysAuth:
				fmt.Printf("Access Key: %s...\n", auth.AccessKeyID[:min(10, len(auth.AccessKeyID))])
			}
		case *config.LocalProvider:
			// No additional details for local provider
		}

		changeLine = true
	}

	return nil
}
