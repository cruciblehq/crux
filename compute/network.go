package compute

import "context"

// Configuration for creating a network.
type NetworkOptions struct {
	Subnet string `codec:"subnet,omitempty"` // CIDR notation (e.g. "10.0.1.0/24"); empty means provider-assigned.
}

// Describes an existing network.
type NetworkInfo struct {
	ID     string `codec:"id"`               // Provider-assigned identifier.
	Subnet string `codec:"subnet,omitempty"` // Assigned CIDR.
}

// Manages provider-level network topology and firewall rules.
//
// Instances assigned to different networks have no connectivity. An instance
// can participate in multiple networks, enabling topologies where two peers
// of the same instance cannot reach each other. Firewall rules are applied per
// instance and enforce the deny-all baseline defined by the affordance model.
// The provider assigns a unique identifier to each network at creation time.
type Network interface {

	// Creates a new network and returns its provider-assigned identifier.
	//
	// The network exists independently of any instance. Returns an error if
	// the provider cannot create the network.
	Create(ctx context.Context, opts NetworkOptions) (string, error)

	// Permanently removes the network.
	//
	// No instances may be attached to the network. Returns an error if the
	// network does not exist or still has attached instances.
	Remove(ctx context.Context, id string) error

	// Attaches an instance to a network.
	//
	// The instance gains L2/L3 reachability to other instances on the same
	// network. An instance may be attached to multiple networks. Returns an
	// error if either identifier is invalid or if the instance is already
	// attached to this network.
	Attach(ctx context.Context, instance, network string) error

	// Detaches an instance from a network.
	//
	// The instance loses reachability to other instances on that network.
	// Returns an error if the instance is not attached to the network.
	Detach(ctx context.Context, instance, network string) error

	// Replaces the firewall rules for an instance.
	//
	// instance is the identifier returned by [Compute.Provision]. rules is the
	// complete set of allow rules derived from the instance's affordance
	// declarations. The provider translates them into its native firewall
	// mechanism: nftables for the local provider, security groups for cloud
	// providers. Calling Apply replaces any previously applied rules for the
	// instance. An empty slice restores the deny-all baseline.
	Apply(ctx context.Context, instance string, rules []Rule) error

	// Returns all networks known to the provider.
	//
	// Returns an empty slice when no networks have been created.
	List(ctx context.Context) ([]NetworkInfo, error)
}
