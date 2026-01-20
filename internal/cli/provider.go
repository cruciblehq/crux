package cli

// Manages cloud provider configurations.
type ProviderCmd struct {
	Add        *ProviderAddCmd        `cmd:"" help:"Add a new cloud provider configuration"`
	List       *ProviderListCmd       `cmd:"" help:"List configured providers"`
	Remove     *ProviderRemoveCmd     `cmd:"" help:"Remove a provider configuration"`
	SetDefault *ProviderSetDefaultCmd `cmd:"" help:"Set the default provider"`
}
