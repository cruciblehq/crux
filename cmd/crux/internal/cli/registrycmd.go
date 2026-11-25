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
	"github.com/cruciblehq/spec/registry"
	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
	"golang.org/x/term"
)

// Registry settings.
const (
	builtinRegistry      = "crucible"        // Registry that is always available, at [internal.DefaultRegistryURL].
	registriesFile       = "registries.yaml" // File in the config directory that lists the configured registries.
	registrySecretPrefix = "registry."       // Prefix of the file name in the secrets directory holding a registry's credential.
)

// Description of errors from reading or writing the registry settings.
const descRegistrySettings = "cannot use registry settings"

// Registry names: lowercase letters, digits, and hyphens, starting with a letter.
var registryNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// Manages the registries crux pushes to and pulls from.
type RegistryCmd struct {
	Add    RegistryAddCmd    `cmd:"" help:"Add a registry, or store the credential for one."`
	List   RegistryListCmd   `cmd:"" aliases:"ls" help:"List the registries."`
	Remove RegistryRemoveCmd `cmd:"" aliases:"rm" help:"Remove a registry and its credential."`
	Use    RegistryUseCmd    `cmd:"" help:"Set the default registry."`
}

// Registry crux knows about.
type registryEntry struct {
	Name string `codec:"name"` // Name the registry is referred to by.
	URL  string `codec:"url"`  // Base URL of the registry.
}

// Registries the user configured, as kept in [registriesFile].
type registryConfig struct {
	Default    string          `codec:"default,omitempty"`    // Name of the default registry; empty for [builtinRegistry].
	Registries []registryEntry `codec:"registries,omitempty"` // Configured registries other than [builtinRegistry].
}

// Reads the registry settings, or returns empty settings if there are none yet.
func loadRegistryConfig() (*registryConfig, error) {
	data, err := os.ReadFile(registryConfigPath())
	if errors.Is(err, fs.ErrNotExist) {
		return &registryConfig{}, nil
	}
	if err != nil {
		return nil, registryConfigError(err)
	}
	var c registryConfig
	if err := codec.Decode(data, &c, codec.YAML); err != nil {
		return nil, registryConfigError(err)
	}
	return &c, nil
}

// Writes the registry settings.
func (c *registryConfig) save() error {
	data, err := codec.Encode(c, codec.YAML)
	if err != nil {
		return registryConfigError(err)
	}
	if err := os.MkdirAll(dirs.Config(), file.DefaultDirMode); err != nil {
		return registryConfigError(err)
	}
	if err := file.WriteFile(registryConfigPath(), data, file.DefaultFileMode); err != nil {
		return registryConfigError(err)
	}
	return nil
}

// Returns the registry called name, including [builtinRegistry].
func (c *registryConfig) find(name string) (registryEntry, bool) {
	if name == builtinRegistry {
		return registryEntry{Name: builtinRegistry, URL: internal.DefaultRegistryURL}, true
	}
	for _, e := range c.Registries {
		if e.Name == name {
			return e, true
		}
	}
	return registryEntry{}, false
}

// Returns every registry, starting with [builtinRegistry].
func (c *registryConfig) all() []registryEntry {
	builtin, _ := c.find(builtinRegistry)
	return append([]registryEntry{builtin}, c.Registries...)
}

// Adds e, replacing a registry with the same name.
func (c *registryConfig) set(e registryEntry) {
	for i := range c.Registries {
		if c.Registries[i].Name == e.Name {
			c.Registries[i] = e
			return
		}
	}
	c.Registries = append(c.Registries, e)
}

// Removes the registry called name, and makes [builtinRegistry] the default if it was.
func (c *registryConfig) remove(name string) {
	c.Registries = slices.DeleteFunc(c.Registries, func(e registryEntry) bool {
		return e.Name == name
	})
	if c.Default == name {
		c.Default = ""
	}
}

// Returns the name of the default registry.
func (c *registryConfig) defaultName() string {
	if c.Default == "" {
		return builtinRegistry
	}
	return c.Default
}

// Returns the path of the registry settings file.
func registryConfigPath() string {
	return filepath.Join(dirs.Config(), registriesFile)
}

// Returns the path of the file holding the credential of the registry called name.
func registryCredentialPath(name string) string {
	return filepath.Join(dirs.Secrets(), registrySecretPrefix+name)
}

// Returns the credential stored for the registry called name, or an empty string.
func registryCredential(name string) (string, error) {
	credential, err := os.ReadFile(registryCredentialPath(name))
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", crex.SystemError(descRegistrySettings, "the registry credential could not be read").
			Recoveryf("Check the permissions of %s, then try again.", dirs.Secrets()).
			Cause(err).
			Err()
	}
	return string(credential), nil
}

// Returns a source for the registry called name, with its credential.
//
// An empty name selects the default registry.
func openRegistrySource(name string) (registry.Source, error) {
	c, err := loadRegistryConfig()
	if err != nil {
		return registry.Source{}, err
	}
	if name == "" {
		name = c.defaultName()
	}
	entry, ok := c.find(name)
	if !ok {
		return registry.Source{}, unknownRegistryError(name)
	}

	src, err := registry.NewSource(entry.URL, internal.DefaultNamespace, dirs.Registry())
	if err != nil {
		return registry.Source{}, err
	}
	if src.Credential, err = registryCredential(name); err != nil {
		return registry.Source{}, err
	}
	return src, nil
}

// Checks that name can name a registry.
func validateRegistryName(name string) error {
	if registryNamePattern.MatchString(name) {
		return nil
	}
	return crex.UserErrorf("invalid registry name", "registry name %q must start with a lowercase letter and contain only lowercase letters, digits, and hyphens", name).
		Recovery("Choose a name such as 'staging', then try again.").
		Err()
}

// Checks that raw is an http or https URL with a host.
func validateRegistryURL(raw string) error {
	u, err := url.Parse(raw)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
		return nil
	}
	return crex.UserErrorf("invalid registry URL", "registry URL %q is not an http or https URL with a host", raw).
		Recovery("Use a URL such as 'https://registry.example.com', then try again.").
		Err()
}

// Returns the error for a registry name that is not configured.
func unknownRegistryError(name string) error {
	return crex.UserErrorf("unknown registry", "no registry named %q is configured", name).
		Recovery("Run 'crux registry list' to see the registries, or 'crux registry add' to add one.").
		Err()
}

// Returns the error for a registry settings file that cannot be read or written.
func registryConfigError(err error) error {
	return crex.SystemError(descRegistrySettings, "the registry settings file could not be read or written").
		Recoveryf("Check %s, then try again.", registryConfigPath()).
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
