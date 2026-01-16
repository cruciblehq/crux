package cli

// Manages cloud provider configurations.
type ProviderCmd struct {
	Add        *ProviderAddCmd        `cmd:"" help:"Add a new cloud provider configuration"` // Add a new provider
	List       *ProviderListCmd       `cmd:"" help:"List configured providers"`              // List providers
	Remove     *ProviderRemoveCmd     `cmd:"" help:"Remove a provider configuration"`        // Remove a provider
	SetDefault *ProviderSetDefaultCmd `cmd:"" help:"Set the default provider"`               // Set the default provider
}
