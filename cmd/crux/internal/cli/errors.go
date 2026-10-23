package cli

import (
	"errors"

	"github.com/cruciblehq/utils-go/crex"
)

var (
	ErrUnexpectedState = crex.New("unexpected runtime state")
	ErrImport          = crex.New("import failed")
)

// Returns err unchanged when it is already structured (crex), otherwise, wraps
// it in a generic CLI error.
func ensureCLIError(err error) error {
	if err == nil {
		return nil
	}
	var cerr *crex.Error
	if errors.As(err, &cerr) {
		return err
	}
	return crex.SystemError("command failed", "an unexpected internal error occurred").
		Recovery("If the problem persists, report it to the Crucible team.").
		Cause(err).
		Err()
}
