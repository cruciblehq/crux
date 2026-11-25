package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/cruciblehq/crux/cmd/crux/internal"
	"github.com/cruciblehq/crux/dirs"
	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
	"golang.org/x/term"
)

// Hub settings.
const (
	builtinHub      = "crucible"  // Hub that is always available, at [internal.DefaultRegistryURL].
	hubsFile        = "hubs.yaml" // File in the config directory that lists the configured hubs.
	hubSecretPrefix = "hub."      // Prefix of the file name in the secrets directory holding a hub's credential.
)

// Description of errors from reading or writing the hub settings.
const descHubSettings = "cannot use hub settings"

// Hub names: lowercase letters, digits, and hyphens, starting with a letter.
var hubNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// Manages the hubs crux pushes to and pulls from.
type HubCmd struct {
	Add    HubAddCmd    `cmd:"" help:"Add a hub, or store the credential for one."`
	List   HubListCmd   `cmd:"" aliases:"ls" help:"List the hubs."`
	Remove HubRemoveCmd `cmd:"" aliases:"rm" help:"Remove a hub and its credential."`
	Use    HubUseCmd    `cmd:"" help:"Set the default hub."`
}

// Hub crux knows about.
type hubEntry struct {
	Name string `codec:"name"` // Name the hub is referred to by.
	URL  string `codec:"url"`  // Base URL of the hub's registry API.
}

// Hubs the user configured, as kept in [hubsFile].
type hubConfig struct {
	Default string     `codec:"default,omitempty"` // Name of the default hub; empty for [builtinHub].
	Hubs    []hubEntry `codec:"hubs,omitempty"`    // Configured hubs other than [builtinHub].
}

// Reads the hub settings, or returns empty settings if there are none yet.
func loadHubConfig() (*hubConfig, error) {
	data, err := os.ReadFile(hubConfigPath())
	if errors.Is(err, fs.ErrNotExist) {
		return &hubConfig{}, nil
	}
	if err != nil {
		return nil, hubConfigError(err)
	}
	var c hubConfig
	if err := codec.Decode(data, &c, codec.YAML); err != nil {
		return nil, hubConfigError(err)
	}
	return &c, nil
}

// Writes the hub settings.
func (c *hubConfig) save() error {
	data, err := codec.Encode(c, codec.YAML)
	if err != nil {
		return hubConfigError(err)
	}
	if err := os.MkdirAll(dirs.Config(), file.DefaultDirMode); err != nil {
		return hubConfigError(err)
	}
	if err := file.WriteFile(hubConfigPath(), data, file.DefaultFileMode); err != nil {
		return hubConfigError(err)
	}
	return nil
}

// Returns the hub called name, including [builtinHub].
func (c *hubConfig) find(name string) (hubEntry, bool) {
	if name == builtinHub {
		return hubEntry{Name: builtinHub, URL: internal.DefaultRegistryURL}, true
	}
	for _, e := range c.Hubs {
		if e.Name == name {
			return e, true
		}
	}
	return hubEntry{}, false
}

// Returns every hub, starting with [builtinHub].
func (c *hubConfig) all() []hubEntry {
	builtin, _ := c.find(builtinHub)
	return append([]hubEntry{builtin}, c.Hubs...)
}

// Adds e, replacing a hub with the same name.
func (c *hubConfig) set(e hubEntry) {
	for i := range c.Hubs {
		if c.Hubs[i].Name == e.Name {
			c.Hubs[i] = e
			return
		}
	}
	c.Hubs = append(c.Hubs, e)
}

// Removes the hub called name, and makes [builtinHub] the default if it was.
func (c *hubConfig) remove(name string) {
	c.Hubs = slices.DeleteFunc(c.Hubs, func(e hubEntry) bool {
		return e.Name == name
	})
	if c.Default == name {
		c.Default = ""
	}
}

// Returns the name of the default hub.
func (c *hubConfig) defaultName() string {
	if c.Default == "" {
		return builtinHub
	}
	return c.Default
}

// Returns the path of the hub settings file.
func hubConfigPath() string {
	return filepath.Join(dirs.Config(), hubsFile)
}

// Returns the path of the file holding the credential of the hub called name.
func hubCredentialPath(name string) string {
	return filepath.Join(dirs.Secrets(), hubSecretPrefix+name)
}

// Returns the credential stored for the hub called name, or an empty string.
func hubCredential(name string) (string, error) {
	credential, err := os.ReadFile(hubCredentialPath(name))
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", crex.SystemError(descHubSettings, "the hub credential could not be read").
			Recoveryf("Check the permissions of %s, then try again.", dirs.Secrets()).
			Cause(err).
			Err()
	}
	return string(credential), nil
}

// Returns a source for the hub called name, with its credential.
//
// An empty name selects the default hub.
func openHubSource(name string) (hub.Source, error) {
	c, err := loadHubConfig()
	if err != nil {
		return hub.Source{}, err
	}
	if name == "" {
		name = c.defaultName()
	}
	entry, ok := c.find(name)
	if !ok {
		return hub.Source{}, unknownHubError(name)
	}

	src, err := hub.NewSource(entry.URL, internal.DefaultNamespace)
	if err != nil {
		return hub.Source{}, err
	}
	if src.Credential, err = hubCredential(name); err != nil {
		return hub.Source{}, err
	}
	return src, nil
}

// Checks that name can name a hub.
func validateHubName(name string) error {
	if hubNamePattern.MatchString(name) {
		return nil
	}
	return crex.UserErrorf("invalid hub name", "hub name %q must start with a lowercase letter and contain only lowercase letters, digits, and hyphens", name).
		Recovery("Choose a name such as 'staging', then try again.").
		Err()
}

// Checks that raw is an http or https URL with a host.
func validateHubURL(raw string) error {
	u, err := url.Parse(raw)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
		return nil
	}
	return crex.UserErrorf("invalid hub URL", "hub URL %q is not an http or https URL with a host", raw).
		Recovery("Use a URL such as 'https://hub.example.com', then try again.").
		Err()
}

// Returns the error for a hub name that is not configured.
func unknownHubError(name string) error {
	return crex.UserErrorf("unknown hub", "no hub named %q is configured", name).
		Recovery("Run 'crux hub list' to see the hubs, or 'crux hub add' to add one.").
		Err()
}

// Returns the error for a hub settings file that cannot be read or written.
func hubConfigError(err error) error {
	return crex.SystemError(descHubSettings, "the hub settings file could not be read or written").
		Recoveryf("Check %s, then try again.", hubConfigPath()).
		Cause(err).
		Err()
}

// Reads a credential from standard input.
//
// On a terminal, prompt is shown and the input is not echoed. Otherwise the
// first line of standard input is read, so a credential can be piped in.
func readCredential(prompt string) (string, error) {
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		fmt.Fprint(os.Stderr, prompt)
		b, err := term.ReadPassword(fd)
		fmt.Fprintln(os.Stderr)
		return strings.TrimSpace(string(b)), err
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
