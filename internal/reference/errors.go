package reference

import (
	"errors"
	"fmt"
)

var (

	// Broad sentinel errors.

	ErrInvalidIdentifier = errors.New("invalid identifier")
	ErrInvalidReference  = errors.New("invalid reference")
	ErrInvalidVersion    = errors.New("invalid version")
	ErrInvalidDigest     = errors.New("invalid digest")
	ErrTypeMismatch      = errors.New("resource type mismatch")

	// Identifier errors.

	ErrInvalidContextType = errors.New("invalid context type")
	ErrEmptyIdentifier    = errors.New("empty identifier")
	ErrInvalidRegistry    = errors.New("invalid registry")
	ErrInvalidPath        = errors.New("invalid path")
	ErrInvalidNamespace   = errors.New("invalid namespace")
	ErrInvalidName        = errors.New("invalid name")

	// Reference errors.

	ErrEmptyReference        = errors.New("empty reference")
	ErrMissingVersionChannel = errors.New("missing version or channel")

	// Version constraint errors.

	ErrEmptyConstraint           = errors.New("empty constraint string")
	ErrEmptyConstraintGroup      = errors.New("empty constraint group")
	ErrBareWildcard              = errors.New("bare wildcard not allowed")
	ErrMultipleWildcards         = errors.New("multiple wildcards not allowed")
	ErrWildcardWithOperator      = errors.New("wildcard cannot have operator")
	ErrPrereleaseInConstraint    = errors.New("prerelease not allowed in constraint")
	ErrLeadingHyphen             = errors.New("leading hyphen in range")
	ErrTrailingHyphen            = errors.New("trailing hyphen in range")
	ErrConsecutiveHyphens        = errors.New("consecutive hyphens in range")
	ErrHyphenWithOperator        = errors.New("hyphen range with operator")
	ErrRangeBoundWithOperator    = errors.New("range bound cannot have operator")
	ErrRangeBoundWithWildcard    = errors.New("range bound cannot have wildcard")
	ErrMissingUpperBound         = errors.New("constraint requires explicit upper bound")
	ErrInvalidVersionFormat      = errors.New("invalid version format")
	ErrInvalidConstraintOperator = errors.New("invalid constraint operator")
	ErrInvalidRangeBound         = errors.New("invalid range bound")
	ErrEmptyOrExpression         = errors.New("empty version constraint in OR expression")
	ErrNilConstraint             = errors.New("cannot intersect nil constraints")
	ErrIncompatibleConstraints   = errors.New("constraints have no common versions")
	ErrUnexpectedToken           = errors.New("unexpected token")

	// Version errors.

	ErrInvalidBuildMetadata     = errors.New("invalid build metadata")
	ErrInvalidPrereleaseFormat  = errors.New("invalid prerelease format")
	ErrInvalidVersionComponents = errors.New("version must have major.minor.patch")
	ErrInvalidMajorVersion      = errors.New("invalid major version")
	ErrInvalidMinorVersion      = errors.New("invalid minor version")
	ErrInvalidPatchVersion      = errors.New("invalid patch version")

	// Digest errors.

	ErrMissingDigestColon   = errors.New("digest missing colon separator")
	ErrEmptyDigestAlgorithm = errors.New("empty digest algorithm")
	ErrEmptyDigestHash      = errors.New("empty digest hash")
)

// Wraps two errors into one.
//
// The sentinel error should be a broad category (e.g., ErrInvalidReference)
// and the wrapped error should provide specific details (e.g.,
// ErrInvalidVersionFormat). This allows callers to check for the broad
// category while still retaining access to the specific error information.
//
// Note: I use this mostly to shut linters up about wrapping errors with
// fmt.Errorf without losing the original error type.
func wrap(sentinel, err error) error {
	return fmt.Errorf("%w: %w", sentinel, err)
}
