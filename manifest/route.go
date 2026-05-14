package manifest

// A routing rule in the gateway.
//
// A route binds an incoming URL path pattern to a service. When a request
// arrives at the gateway and its path exactly matches [Route.Pattern], the
// gateway forwards it to the service identified by [Route.Service]. Matching
// is exact: the pattern "/foo" matches the path "/foo" but not "/foobar".
type Route struct {

	// URL path that triggers this route (e.g. "/api", "/auth").
	//
	// Must be non-empty and unique across all routes in the gateway. Matching
	// is exact: two paths are equal only if every character is identical.
	// Two routes with the same pattern are an error.
	Pattern string `codec:"pattern"`

	// ID of the service that receives requests matched by this route.
	//
	// Must refer to a service declared in [Blueprint.Services]. Services that
	// are not referenced by any route still run but receive no external traffic.
	Service string `codec:"service"`
}

// Validates the route.
//
// The pattern must be a valid URL path: starting with "/", no consecutive
// slashes, no query strings or fragments, unreserved characters only. The
// service ID must be a valid name per [isValidName].
func (r *Route) Validate() error {
	if !isValidRoutePattern(r.Pattern) {
		return ErrInvalidRoutePattern
	}
	if !isValidName(r.Service) {
		return ErrInvalidRouteService
	}
	return nil
}
