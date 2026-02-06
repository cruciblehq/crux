package reference

import "errors"

var (

	// Broad sentinel errors
	ErrInvalidIdentifier = errors.New("invalid identifier")
	ErrInvalidReference  = errors.New("invalid reference")
	ErrInvalidDigest     = errors.New("invalid digest")
	ErrTypeMismatch      = errors.New("resource type mismatch")

	// Specific identifier errors
	ErrInvalidContextType    = errors.New("invalid context type")
	ErrEmptyIdentifier       = errors.New("empty identifier")
	ErrExtraIdentifierTokens = errors.New("extra tokens in identifier")
	ErrInvalidScheme         = errors.New("invalid scheme")
	ErrInvalidRegistry       = errors.New("invalid registry")
	ErrInvalidPath           = errors.New("invalid path")
	ErrInvalidNamespace      = errors.New("invalid namespace")
	ErrInvalidName           = errors.New("invalid name")
	ErrMissingRegistry       = errors.New("missing registry in URI")
	ErrMissingPath           = errors.New("missing path in URI")
	ErrEmptyPath             = errors.New("empty path")

	// Specific reference errors
	ErrEmptyReference        = errors.New("empty reference")
	ErrExtraReferenceTokens  = errors.New("extra tokens in reference")
	ErrNoIdentifier          = errors.New("no identifier found")
	ErrMissingVersionChannel = errors.New("missing version or channel")

	// Specific version constraint errors
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

	// Specific version errors
	ErrInvalidBuildMetadata     = errors.New("invalid build metadata")
	ErrInvalidPrereleaseFormat  = errors.New("invalid prerelease format")
	ErrInvalidVersionComponents = errors.New("version must have major.minor.patch")
	ErrInvalidMajorVersion      = errors.New("invalid major version")
	ErrInvalidMinorVersion      = errors.New("invalid minor version")
	ErrInvalidPatchVersion      = errors.New("invalid patch version")
	ErrNegativeMajorVersion     = errors.New("negative major version")
	ErrNegativeMinorVersion     = errors.New("negative minor version")
	ErrNegativePatchVersion     = errors.New("negative patch version")

	// Specific digest errors
	ErrMissingDigestColon   = errors.New("digest missing colon separator")
	ErrEmptyDigestAlgorithm = errors.New("empty digest algorithm")
	ErrEmptyDigestHash      = errors.New("empty digest hash")
)
