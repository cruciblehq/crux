package manifest

// A named set of environment variables.
//
// Environments provide concrete values for affordances declared by services.
// The specific environment to use is selected at build time, impacting the
// resulting deployment plan.
type Environment struct {

	// Unique identifier for this environment.
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
