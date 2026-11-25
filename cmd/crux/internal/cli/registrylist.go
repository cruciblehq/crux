package cli

import (
	"context"
	"fmt"
)

// Represents the 'crux registry list' command.
type RegistryListCmd struct {
}

// Prints each registry with its URL, marking the default and registries without a
// credential.
func (c *RegistryListCmd) Run(_ context.Context) error {
	cfg, err := loadRegistryConfig()
	if err != nil {
		return err
	}

	defaultName := cfg.defaultName()
	for _, e := range cfg.all() {
		credential, err := registryCredential(e.Name)
		if err != nil {
			return err
		}
		notes := ""
		if e.Name == defaultName {
			notes += " (default)"
		}
		if credential == "" {
			notes += " (no credential)"
		}
		fmt.Printf("%s %s%s\n", e.Name, e.URL, notes)
	}
	return nil
}
