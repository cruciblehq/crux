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

// Represents the 'crux hub add' command.
type HubAddCmd struct {
	Name string `arg:"" help:"Name to refer to the hub by."`
	URL  string `arg:"" optional:"" help:"Base URL of the hub. Not needed for the built-in 'crucible' hub."`
}

// Adds a hub and stores its credential.
//
// For the built-in hub only the credential is stored. The credential is read
// from standard input, without echo on a terminal.
func (c *HubAddCmd) Run(_ context.Context) error {
	if err := validateHubName(c.Name); err != nil {
		return err
	}
	if err := c.validateURL(); err != nil {
		return err
	}

	cfg, err := loadHubConfig()
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
			Recovery("Run 'crux hub add' again and enter the credential for the hub.").
			Err()
	}
	err = os.MkdirAll(dirs.Secrets(), file.OwnerDirMode)
	if err == nil {
		err = file.WriteFile(hubCredentialPath(c.Name), []byte(credential), file.OwnerFileMode)
	}
	if err != nil {
		return crex.SystemError("cannot store credential", "the hub credential could not be stored").
			Recoveryf("Check the permissions of %s, then try again.", dirs.Secrets()).
			Cause(err).
			Err()
	}

	if c.Name != builtinHub {
		cfg.set(hubEntry{Name: c.Name, URL: c.URL})
		if err := cfg.save(); err != nil {
			return err
		}
	}

	slog.Info("hub added", "name", c.Name)
	return nil
}

// Checks the URL against the hub name.
//
// The built-in hub takes no URL; every other hub needs one.
func (c *HubAddCmd) validateURL() error {
	if c.Name == builtinHub {
		if c.URL != "" {
			return crex.UserErrorf("unexpected hub URL", "the %q hub is built in and has a fixed URL", builtinHub).
				Recoveryf("Run 'crux hub add %s' without a URL.", builtinHub).
				Err()
		}
		return nil
	}
	if c.URL == "" {
		return crex.UserErrorf("missing hub URL", "hub %q needs a URL", c.Name).
			Recoveryf("Run 'crux hub add %s <url>'.", c.Name).
			Err()
	}
	return validateHubURL(c.URL)
}
