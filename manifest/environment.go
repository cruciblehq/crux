package manifest

// A named set of environment variables.
//
// Environments provide concrete values for the config/env and config/secret
// affordances declared by services. Which environment to use is selected a
// build time, producing different plans from the same blueprint.
type Environment struct {

	// Unique identifier for this environment (e.g. "production", "staging").
	ID string `codec:"id"`

	// Key-value pairs for this environment.
	Variables map[string]string `codec:"variables"`
}

// Validates the environment entry.
func (e *Environment) Validate() error {
	if e.ID == "" {
		return ErrMissingEnvironmentID
	}
	if !isValidName(e.ID) {
		return ErrInvalidEnvironmentID
	}
	return nil
}
