package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/paths"
	"github.com/cruciblehq/protocol/pkg/codec"
)

// ProviderType specifies the type of cloud provider.
type ProviderType string

const (

	// ProviderTypeAWS represents Amazon Web Services.
	ProviderTypeAWS ProviderType = "aws"

	// ProviderTypeLocal represents local Docker-based deployment.
	ProviderTypeLocal ProviderType = "local"
)

// Cast a string to a ProviderType.
//
// Returns an error if the string does not correspond to a valid ProviderType.
func ProviderTypeFromString(s string) (ProviderType, error) {
	switch s {
	case "aws":
		return ProviderTypeAWS, nil
	case "local":
		return ProviderTypeLocal, nil
	default:
		return "", ErrInvalidProviderType
	}
}

// Stores user's configured cloud providers.
//
// The configuration is stored in the user's config directory (i.e., at
// paths.Providers()). The file is created when the first provider is added.
// If no configuration file exists, an empty configuration is returned by
// [LoadProviders]. The Default field indicates the name of the default provider,
// which can be set using [SetDefault] and retrieved using [GetDefault]. This
// field should not be modified directly, use the provided methods instead.
type ProvidersConfig struct {
	Default   string              `field:"default"`   // Name of the default provider
	Providers map[string]Provider `field:"providers"` // Map of provider name to Provider config
}

// Represents a cloud provider configuration.
//
// Config is a polymorphic field and its concrete type depends on the value of
// Type (e.g., *[AWSProvider] for ProviderTypeAWS and *[LocalProvider] for
// ProviderTypeLocal).
type Provider struct {
	Type   ProviderType `field:"type"`             // Provider type (ProviderTypeAWS, ProviderTypeLocal)
	Name   string       `field:"name"`             // Name of this provider configuration
	Config any          `field:"config,omitempty"` // Provider-specific configuration
}

// Reads the providers configuration.
//
// The configuration is read from the user's config directory (i.e., at
// paths.Providers()). If the configuration file does not exist, an empty
// configuration is returned.
func LoadProviders() (*ProvidersConfig, error) {
	configPath := paths.Providers()

	// Load existing config
	var config ProvidersConfig
	if _, err := codec.DecodeFile(configPath, "field", &config); err != nil {

		// If file doesn't exist, return empty config
		if errors.Is(err, os.ErrNotExist) {
			return &ProvidersConfig{
				Providers: make(map[string]Provider),
			}, nil
		}

		return nil, crex.Wrap(ErrInvalidProvider, err)
	}

	// The configuration file exists but has no providers; return empty map
	if config.Providers == nil {
		config.Providers = make(map[string]Provider)
	}

	return &config, nil
}

// Writes the providers configuration to disk.
//
// The configuration is written to the user's config directory (i.e., at
// paths.Providers()). If the configuration file's parent directory does not
// exist, it is created. If the file already exists, it is overwritten.
func (c *ProvidersConfig) Save() error {
	configPath := paths.Providers()

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, paths.DefaultDirMode); err != nil {
		return err
	}

	// Write config with restricted permissions (user only)
	if err := codec.EncodeFile(configPath, "field", c); err != nil {
		return err
	}

	return nil
}

// Adds or updates a provider configuration.
//
// If a provider with the same name already exists, it is overwritten. If this
// is the first provider being added, it is set as the default provider. The
// changes are not persisted to disk until [Save] is called.
func (c *ProvidersConfig) AddProvider(name string, provider Provider) {
	provider.Name = name
	c.Providers[name] = provider

	// If this is the first provider, make it default
	if c.Default == "" {
		c.Default = name
	}
}

// Removes a provider configuration.
//
// If the provider does not exist, an error is returned. If the removed provider
// was the default, the default is cleared (no default). The changes are not
// persisted to disk until [Save] is called.
func (c *ProvidersConfig) RemoveProvider(name string) error {
	if _, exists := c.Providers[name]; !exists {
		return ErrProviderNotFound
	}

	delete(c.Providers, name)

	// If we removed the default, clear it
	if c.Default == name {
		c.Default = ""
	}

	return nil
}

// Gets a provider by name.
//
// Returns the provider configuration or an error if the provider does not exist.
func (c *ProvidersConfig) GetProvider(name string) (Provider, error) {
	provider, exists := c.Providers[name]
	if !exists {
		return Provider{}, ErrProviderNotFound
	}
	return provider, nil
}

// Gets the default provider.
//
// Returns the default provider configuration or an error if no default is set.
// The provider with the given name must exist, otherwise an error is returned.
func (c *ProvidersConfig) GetDefault() (Provider, error) {
	if c.Default == "" {
		return Provider{}, ErrProviderNotFound
	}
	return c.GetProvider(c.Default)
}

// Sets the default provider.
//
// The provider with the given name must exist, otherwise an error is returned.
// If an existing default is set, it will be overriden.
func (c *ProvidersConfig) SetDefault(name string) error {
	if _, exists := c.Providers[name]; !exists {
		return ErrProviderNotFound
	}
	c.Default = name
	return nil
}

// Gets a provider by name, or the default provider if name is empty.
//
// If name is empty and no default is set, returns an error. If name is provided
// but the provider does not exist, returns an error.
func (c *ProvidersConfig) GetOrDefault(name string) (Provider, error) {
	if name == "" {
		return c.GetDefault()
	}
	return c.GetProvider(name)
}

// Returns a list of configured providers.
//
// Returns a copy of the providers slice or an empty slice if no provider is
// configured. Order is not guaranteed.
func (c *ProvidersConfig) ListProviders() []Provider {
	providers := make([]Provider, 0, len(c.Providers))
	for _, provider := range c.Providers {
		providers = append(providers, provider)
	}
	return providers
}
