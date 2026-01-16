package cli

import (
	"bufio"
	"context"
	"log/slog"
	"os"

	"github.com/cruciblehq/crux/pkg/config"
	"github.com/cruciblehq/crux/pkg/crex"
)

const (

	// Default AWS region
	DefaultAWSRegion = "us-east-1"
)

// Add a new cloud provider configuration.
type ProviderAddCmd struct {
	Name string `arg:"" help:"Name for this provider configuration (e.g., aws-production)"` // Name of the provider
}

// Run executes the provider add command.
//
// Prompts the user for provider information (type, region, credentials, etc.),
// validates the input, and saves the configuration to disk. If this is the first
// provider being added, it automatically becomes the default provider.
func (c *ProviderAddCmd) Run(ctx context.Context) error {
	cfg, err := config.LoadProviders()
	if err != nil {
		return err
	}

	provider, err := promptProviderInfo()
	if err != nil {
		return err
	}

	cfg.AddProvider(c.Name, *provider)

	if err := cfg.Save(); err != nil {
		return err
	}

	slog.Info("Provider added successfully", "name", c.Name, "type", provider.Type)

	return nil
}

// Prompts the user for provider information and returns a configured provider.
//
// Asks the user to select a provider type (AWS or local) and then gathers the
// necessary configuration details for that provider type. Returns a fully
// configured Provider struct or an error if the input is invalid.
func promptProviderInfo() (*config.Provider, error) {
	reader := bufio.NewReader(os.Stdin)

	providerTypeString := promptWithDefault(reader, "Provider type", "local")
	providerType, err := config.ProviderTypeFromString(providerTypeString)
	if err != nil {
		return nil, err
	}

	switch providerType {
	case config.ProviderTypeAWS:
		awsConfig, err := promptAWSProviderInfo(reader)
		if err != nil {
			return nil, err
		}
		return &config.Provider{
			Type:   providerType,
			Config: awsConfig,
		}, nil
	case config.ProviderTypeLocal:
		return &config.Provider{
			Type:   providerType,
			Config: &config.LocalProvider{},
		}, nil
	default:
		return nil, crex.UserErrorf("invalid provider type", "unsupported provider type: %s", providerTypeString).Err()
	}
}

// Prompts the user for AWS provider configuration details.
//
// Collects the AWS region and authentication method (profile or access keys).
// If profile-based authentication is selected, returns immediately with the
// profile configuration. Otherwise, prompts for access key credentials.
func promptAWSProviderInfo(reader *bufio.Reader) (*config.AWSProvider, error) {
	cfg := &config.AWSProvider{}

	// Region
	cfg.Region = promptWithDefault(reader, "Region", DefaultAWSRegion)

	// Profile
	profile := promptWithDefault(reader, "AWS profile name (leave empty to use access keys)", "")

	// Access keys
	if profile != "" {
		cfg.AuthMethod = config.AuthMethodProfile
		cfg.Auth = &config.AWSProfileAuth{
			Profile: profile,
		}
		return cfg, nil
	}

	auth, err := promptAWSAuthKeys(reader)
	if err != nil {
		return nil, err
	}

	cfg.AuthMethod = config.AuthMethodKeys
	cfg.Auth = auth

	return cfg, nil
}

// Prompts the user for AWS access key credentials.
//
// Collects the AWS access key ID and secret access key from the user,
// validates the format and structure of the credentials, and returns a
// configured AWSKeysAuth struct.
func promptAWSAuthKeys(reader *bufio.Reader) (*config.AWSKeysAuth, error) {
	accessKeyID := promptWithDefault(reader, "Access key ID", "")
	secretAccessKey := promptWithDefault(reader, "Secret access key", "")

	auth := &config.AWSKeysAuth{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	}

	if err := auth.Validate(); err != nil {
		return nil, err
	}

	return auth, nil
}
