package manifest

// Gateway routing configuration.
//
// Maps URL path patterns to service IDs. Not every service needs a route;
// services without routes still run but receive no external traffic.
type Gateway struct {
	Routes []Route `codec:"routes,omitempty"` // URL path patterns mapped to services.
}

// Validates gateway routes.
//
// Every route must have a pattern and a service ID. Route patterns must be
// unique. Cross-referencing route service IDs against known services is done
// by [Blueprint.Validate].
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
