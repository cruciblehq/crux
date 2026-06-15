package volume

import "github.com/cruciblehq/crux/crex"

// Persistent storage mount declared by a .volume grant.
//
// The platform provisions the storage and bind-mounts it at Destination before
// the container starts. On local providers this is a directory inside the VM.
// On cloud providers it is a managed persistent disk. The volume is not tied
// to the container lifecycle and survives restarts.
type Mount struct {
	Destination string `codec:"destination"`         // Absolute path inside the container where the volume is mounted.
	ReadOnly    bool   `codec:"read_only,omitempty"` // Whether the volume is mounted read-only.
}

// Validates the mount declaration.
//
// Destination must be a non-empty absolute path.
func (m *Mount) Validate() error {
	if m.Destination == "" {
		return crex.Wrapf(ErrInvalidMount, "destination is empty")
	}
	return nil
}

// Accumulated persistent storage requirements from .volume grants.
//
// Produced by the volume subsystem as each .volume grant is processed. The
// deployer reads Mounts to provision volumes and inject the corresponding bind
// mounts into the OCI runtime spec before starting the container.
type Spec struct {

	// Declared volumes in grant order.
	Mounts []Mount
}

// Validates the volume spec.
func (s *Spec) Validate() error {
	for i := range s.Mounts {
		if err := s.Mounts[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}
