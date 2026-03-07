package resource

import (
	"context"

	"github.com/cruciblehq/spec/protocol"
)

// Narrow interface for the daemon build capability.
//
// Satisfied by [compute.Client]; defined here to avoid an import cycle
// between the resource and compute packages.
type BuildClient interface {

	// Sends a build request to the cruxd daemon.
	//
	// The daemon executes the recipe steps inside a container and produces an
	// OCI image artifact. The result contains the output directory path of the
	// built artifact.
	Build(ctx context.Context, req *protocol.BuildRequest) (*protocol.BuildResult, error)
}
