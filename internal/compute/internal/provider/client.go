package provider

import (
	"context"

	"github.com/cruciblehq/spec/protocol"
)

// Interface for communicating with a cruxd instance.
//
// Each backend implements Client differently depending on how it connects to
// cruxd (Unix socket, SSH tunnel, cloud API, etc.). A Client is obtained
// via [Backend.Client] for a specific instance.
type Client interface {

	// Sends a build request and waits for the result.
	Build(ctx context.Context, req *protocol.BuildRequest) (*protocol.BuildResult, error)

	// Returns the instance's current status.
	Status(ctx context.Context) (*protocol.StatusResult, error)

	// Imports a container image from a local archive.
	ImageImport(ctx context.Context, req *protocol.ImageImportRequest) error

	// Starts a container from a previously imported image.
	ImageStart(ctx context.Context, req *protocol.ImageStartRequest) error

	// Removes a container image and its associated resources.
	ImageDestroy(ctx context.Context, req *protocol.ImageDestroyRequest) error

	// Stops a running container.
	ContainerStop(ctx context.Context, req *protocol.ContainerStopRequest) error

	// Destroys a container and its filesystem state.
	ContainerDestroy(ctx context.Context, req *protocol.ContainerDestroyRequest) error

	// Returns the current state of a container.
	ContainerStatus(ctx context.Context, req *protocol.ContainerStatusRequest) (*protocol.ContainerStatusResult, error)

	// Executes a command inside a running container.
	ContainerExec(ctx context.Context, req *protocol.ContainerExecRequest) (*protocol.ContainerExecResult, error)

	// Updates a running container's configuration.
	ContainerUpdate(ctx context.Context, req *protocol.ContainerUpdateRequest) error
}
