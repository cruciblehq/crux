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

// Represents the 'crux registry remove' command.
type RegistryRemoveCmd struct {
	Name string `arg:"" help:"Name of the registry to remove."`
}

// Removes a registry and its credential.
//
// The built-in registry cannot be removed, so only its credential is deleted.
func (c *RegistryRemoveCmd) Run(_ context.Context) error {
	cfg, err := loadRegistryConfig()
	if err != nil {
		return err
	}
	if _, ok := cfg.find(c.Name); !ok {
		return unknownRegistryError(c.Name)
	}

	if err := os.Remove(registryCredentialPath(c.Name)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return crex.SystemError("cannot remove credential", "the registry credential could not be removed").
			Recoveryf("Check the permissions of %s, then try again.", dirs.Secrets()).
			Cause(err).
			Err()
	}

	cfg.remove(c.Name)
	if err := cfg.save(); err != nil {
		return err
	}

	slog.Info("registry removed", "name", c.Name)
	return nil
}
