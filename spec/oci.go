package spec

import specs "github.com/opencontainers/runtime-spec/specs-go"

// OCI runtime spec type, re-exported from the upstream runtime-spec module.
//
// Aliased here so callers that import this package do not need a separate
// import of the runtime-spec module for container-level types.
type OCI = specs.Spec
