package manifest

import "errors"

var (

	// General errors.

	ErrInvalidResourceType = errors.New("invalid resource type")
	ErrInvalidManifest     = errors.New("invalid manifest")
	ErrInvalidResource     = errors.New("invalid resource")
	ErrInvalidRecipe       = errors.New("invalid recipe")
	ErrInvalidStage        = errors.New("invalid stage")
	ErrInvalidStep         = errors.New("invalid step")
	ErrInvalidService      = errors.New("invalid service")
	ErrInvalidWidget       = errors.New("invalid widget")

	// Manifest.

	ErrUnsupportedVersion = errors.New("unsupported manifest version")
	ErrMissingConfig      = errors.New("missing config")
	ErrConfigTypeMismatch = errors.New("config type does not match resource type")
	ErrEncodeFailed       = errors.New("failed to encode manifest")
	ErrDecodeFailed       = errors.New("failed to decode manifest")

	// Resource.

	ErrMissingName    = errors.New("missing resource name")
	ErrMissingVersion = errors.New("missing resource version")

	// Recipe.

	ErrMissingOutputStage = errors.New("recipe has no output stage")
	ErrDuplicateStageName = errors.New("duplicate stage name")

	// Stage.

	ErrNumericStageName = errors.New("stage name must not be numeric")
	ErrMissingFrom      = errors.New("missing base image")

	// Step.

	ErrMutuallyExclusiveOps    = errors.New("run and copy are mutually exclusive")
	ErrEmptyStep               = errors.New("step has no fields set")
	ErrShellWithCopy           = errors.New("shell cannot be used with copy")
	ErrEnvWithCopy             = errors.New("env cannot be used with copy")
	ErrStepsWithoutPlatform    = errors.New("child steps require platform")
	ErrPlatformWithOperation   = errors.New("platform group cannot have operations")
	ErrNestedPlatformGroup     = errors.New("platform groups cannot be nested")
	ErrPlatformInPlatformStage = errors.New("steps cannot use platform inside a platform-scoped stage")

	// Ref.

	ErrInvalidRef       = errors.New("invalid ref")
	ErrMissingRefTarget = errors.New("missing ref target")
	ErrRefMixed         = errors.New("ref cannot have both a scalar value and args")

	// Affordance.

	ErrInvalidAffordance = errors.New("invalid affordance")

	// Param.

	ErrInvalidParam       = errors.New("invalid param")
	ErrMissingParamName   = errors.New("missing param name")
	ErrDuplicateParamName = errors.New("duplicate param name")
	ErrDefaultNotInSchema = errors.New("default param not in schema")

	// Blueprint.

	ErrInvalidBlueprint = errors.New("invalid blueprint")

	// Service config.

	ErrMissingEntrypoint = errors.New("service missing entrypoint")

	// Widget.

	ErrMissingMain = errors.New("widget missing main entry point")

	// Service refs.

	ErrMissingServiceID   = errors.New("service missing id")
	ErrDuplicateServiceID = errors.New("duplicate service id")

	// Route.

	ErrMissingRoutePattern   = errors.New("route missing pattern")
	ErrMissingRouteService   = errors.New("route missing service")
	ErrRouteServiceNotFound  = errors.New("route references unknown service id")
	ErrDuplicateRoutePattern = errors.New("duplicate route pattern")

	// Environment.

	ErrMissingEnvironmentID   = errors.New("environment missing id")
	ErrDuplicateEnvironmentID = errors.New("duplicate environment id")

	// Plan.

	ErrInvalidPlan             = errors.New("invalid plan")
	ErrUnsupportedPlanVersion  = errors.New("unsupported plan version")
	ErrMissingComputeID        = errors.New("compute missing id")
	ErrMissingProvider         = errors.New("compute missing provider")
	ErrMissingContainerService = errors.New("container missing service")
	ErrMissingContainerCompute = errors.New("container missing compute")

	// State.

	ErrInvalidState            = errors.New("invalid state")
	ErrUnsupportedStateVersion = errors.New("unsupported state version")
	ErrMissingDeployedAt       = errors.New("missing deployment timestamp")
)
