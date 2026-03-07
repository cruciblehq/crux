package provider

// Parameters for provisioning a cruxd runtime instance.
type Config struct {
	Name    string // Identifier for the cruxd instance.
	Version string // Version of cruxd to provision.
}
