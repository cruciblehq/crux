package cli

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/dirs"
	"github.com/cruciblehq/utils-go/crex"
)

// Represents the 'crux hub remove' command.
type HubRemoveCmd struct {
	Name string `arg:"" help:"Name of the hub to remove."`
}

// Removes a hub and its credential.
//
// The built-in hub cannot be removed, so only its credential is deleted.
func (c *HubRemoveCmd) Run(_ context.Context) error {
	cfg, err := loadHubConfig()
	if err != nil {
		return err
	}
	if _, ok := cfg.find(c.Name); !ok {
		return unknownHubError(c.Name)
	}

	if err := os.Remove(hubCredentialPath(c.Name)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return crex.SystemError("cannot remove credential", "the hub credential could not be removed").
			Recoveryf("Check the permissions of %s, then try again.", dirs.Secrets()).
			Cause(err).
			Err()
	}

	cfg.remove(c.Name)
	if err := cfg.save(); err != nil {
		return err
	}

	slog.Info("hub removed", "name", c.Name)
	return nil
}
