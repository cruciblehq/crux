package manifest

// External ingress configuration for a deployment.
//
// The gateway is the entry point for external traffic into a deployment. It
// holds the routing table that maps URL path patterns to services, along with
// any deployment-wide ingress policies that apply before a request reaches a
// service. Not every service needs a route. Internal services such as queues
// and databases run without a route entry and are never reachable from outside
// the deployment.
type Gateway struct {

	// Routing rules, each mapping an exact URL path to a service.
	Routes []Route `codec:"routes,omitempty"`
}

// Validates the gateway and all of its routes.
//
// Every route must have a non-empty pattern and a non-empty service ID.
// Patterns must be unique across the route table. Cross-referencing service
// IDs against the services declared in [Blueprint] is done by
// [Blueprint.Validate], not here.
func (g *Gateway) Validate() error {
	patterns := make(map[string]struct{}, len(g.Routes))
	for i := range g.Routes {
		if err := g.Routes[i].Validate(); err != nil {
			return err
		}
		if _, exists := patterns[g.Routes[i].Pattern]; exists {
			return ErrDuplicateRoutePattern
		}
		patterns[g.Routes[i].Pattern] = struct{}{}
	}
	return nil
}
