package manifest

// A container entry in the deployment plan.
//
// Holds the compiled runtime enforcement spec. Keyed by container ID in the
// parent Plan's Containers map.
type Container struct {
	Runtime *Spec `codec:"runtime,omitempty"` // Compiled enforcement model.
}
