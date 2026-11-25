package cli

import (
	"context"
	"fmt"
)

// Represents the 'crux hub list' command.
type HubListCmd struct {
}

// Prints each hub with its URL, marking the default and hubs without a
// credential.
func (c *HubListCmd) Run(_ context.Context) error {
	cfg, err := loadHubConfig()
	if err != nil {
		return err
	}

	defaultName := cfg.defaultName()
	for _, e := range cfg.all() {
		credential, err := hubCredential(e.Name)
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
