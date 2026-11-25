package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/dirs"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Represents the 'crux registry add' command.
type RegistryAddCmd struct {
	Name string `arg:"" help:"Name to refer to the registry by."`
	URL  string `arg:"" optional:"" help:"Base URL of the registry. Not needed for the built-in 'crucible' registry."`
}

// Adds a registry and stores its credential.
//
// For the built-in registry only the credential is stored. The credential is read
// from standard input, without echo on a terminal.
func (c *RegistryAddCmd) Run(_ context.Context) error {
	if err := validateRegistryName(c.Name); err != nil {
		return err
	}
	if err := c.validateURL(); err != nil {
		return err
	}

	cfg, err := loadRegistryConfig()
	if err != nil {
		return err
	}

	credential, err := readCredential(fmt.Sprintf("Credential for %s: ", c.Name))
	if err != nil {
		return crex.SystemError("cannot read credential", "the credential could not be read from standard input").
			Cause(err).
			Err()
	}
	if credential == "" {
		return crex.UserError("missing credential", "no credential was entered").
			Recovery("Run 'crux registry add' again and enter the credential for the registry.").
			Err()
	}
	err = os.MkdirAll(dirs.Secrets(), file.OwnerDirMode)
	if err == nil {
		err = file.WriteFile(registryCredentialPath(c.Name), []byte(credential), file.OwnerFileMode)
	}
	if err != nil {
		return crex.SystemError("cannot store credential", "the registry credential could not be stored").
			Recoveryf("Check the permissions of %s, then try again.", dirs.Secrets()).
			Cause(err).
			Err()
	}

	if c.Name != builtinRegistry {
		cfg.set(registryEntry{Name: c.Name, URL: c.URL})
		if err := cfg.save(); err != nil {
			return err
		}
	}

	slog.Info("registry added", "name", c.Name)
	return nil
}

// Checks the URL against the registry name.
//
// The built-in registry takes no URL; every other registry needs one.
func (c *RegistryAddCmd) validateURL() error {
	if c.Name == builtinRegistry {
		if c.URL != "" {
			return crex.UserErrorf("unexpected registry URL", "the %q registry is built in and has a fixed URL", builtinRegistry).
				Recoveryf("Run 'crux registry add %s' without a URL.", builtinRegistry).
				Err()
		}
		return nil
	}
	if c.URL == "" {
		return crex.UserErrorf("missing registry URL", "registry %q needs a URL", c.Name).
			Recoveryf("Run 'crux registry add %s <url>'.", c.Name).
			Err()
	}
	return validateRegistryURL(c.URL)
}
