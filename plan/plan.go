package plan

// Current plan format version.
const Version = 0

// Represents a deployment plan.
//
// Specifies what resources will be deployed and the infrastructure configuration
// required to run them. Generated during the planning phase by resolving
// references, allocating infrastructure, and determining routing.
type Plan struct {
	Version      int           `json:"version"`
	Services     []Service     `json:"services"`
	Compute      []Compute     `json:"compute"`
	Environments []Environment `json:"environments,omitempty"`
	Bindings     []Binding     `json:"bindings"`
	Gateway      Gateway       `json:"gateway"`
}

// Represents a service in the deployment plan.
//
// Contains the resolved reference with exact version and digest.
type Service struct {
	ID        string `json:"id"`
	Reference string `json:"reference"`
}

// Represents a compute resource in the deployment plan.
//
// Defines the compute instance to provision. The Config field contains
// provider-specific configuration based on the Provider value.
type Compute struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Config   any    `json:"config,omitempty"`
}

// AWS compute configuration.
//
// Specifies EC2 instance settings for AWS deployments.
type ComputeAWS struct {
	InstanceType string `json:"instance_type"`
	Region       string `json:"region,omitempty"`
}

// Local compute configuration.
//
// No additional configuration needed for local deployments.
type ComputeLocal struct{}

// Represents an environment configuration.
//
// Defines a set of environment variables that can be associated with deployments.
type Environment struct {
	ID        string            `json:"id"`
	Variables map[string]string `json:"variables"`
}

// Represents a binding of a service to compute infrastructure.
//
// Associates a service with a compute instance and optional environment
// configuration.
type Binding struct {
	Service     string `json:"service"`
	Compute     string `json:"compute"`
	Environment string `json:"environment,omitempty"`
}

// Represents the API gateway configuration.
//
// Defines how external requests are routed to deployed services.
type Gateway struct {
	Routes []Route `json:"routes,omitempty"`
}

// Represents a routing rule in the gateway.
//
// Maps request patterns to service instances.
type Route struct {
	Pattern string `json:"pattern"`
	Service string `json:"service"`
}
