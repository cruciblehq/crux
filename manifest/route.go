package manifest

// A gateway route mapping a URL pattern to a service.
type Route struct {
	Pattern string `json:"pattern"` // URL path pattern (e.g. "/api", "/auth").
	Service string `json:"service"` // Service ID that handles requests matching this pattern.
}

// Validates the route.
func (r *Route) Validate() error {
	if r.Pattern == "" {
		return ErrMissingRoutePattern
	}
	if r.Service == "" {
		return ErrMissingRouteService
	}
	return nil
}
