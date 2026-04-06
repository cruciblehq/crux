package manifest

// Identifies the infrastructure provider that the deployment plan targets.
//
// Controls how compute resources are configured in the generated [Plan].
// Each provider has its own configuration schema for compute resources, and
// the provider type determines which schema is used.
type ProviderType string

const (

	// Targets Amazon Web Services. Compute entries in the plan will
	// contain [ComputeAWS] configuration.
	ProviderTypeAWS ProviderType = "aws"

	// Targets the local machine. No additional compute configuration
	// is generated.
	ProviderTypeLocal ProviderType = "local"
)
