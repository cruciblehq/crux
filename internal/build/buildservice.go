package build

import (
	"github.com/cruciblehq/crux/pkg/crex"
	"github.com/cruciblehq/crux/pkg/manifest"
)

// Handles the build process for a service resource.
//
// Services are not yet supported by Crux. It returns an error indicating that
// services require an external build process.
func BuildService(options *manifest.Service) error {
	return crex.UserError("build failed", "services are not built with crux").
		Fallback("Use an external build tool like Docker or Buildah to build your service image.").
		Err()
}
