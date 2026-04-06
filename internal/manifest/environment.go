package manifest

// A named set of environment variables.
//
// Environments provide concrete values for the config/env and config/secret
// affordances declared by services. Which environment to use is selected
// at build time, producing different plans from the same blueprint.
type Environment struct {
	ID        string            `codec:"id"`        // Unique identifier for this environment (e.g. "production", "staging").
	Variables map[string]string `codec:"variables"` // Key-value pairs for this environment.
}

// Validates the environment entry.
func (e *Environment) Validate() error {
	if e.ID == "" {
		return ErrMissingEnvironmentID
	}
	return nil
}
