package manifest

// Holds configuration specific to machine resources.
//
// Machine resources package virtual machine disk images and their metadata so
// they can be distributed through the registry. Build support is intentionally
// separate from the manifest type so images can be packed and published before
// a dedicated builder exists.
type Machine struct{}

// Validates the machine configuration.
func (m *Machine) Validate() error {
	return nil
}
