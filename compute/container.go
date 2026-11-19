package compute

import (
	"context"
	"errors"
	"io"

	"github.com/cruciblehq/spec/agent"
	"github.com/cruciblehq/utils-go/crex"
)

// Make sure the container handle satisfies the agent contract.
var _ agent.Container = (*container)(nil)

// Container on a machine, reached through cruxd over HTTPS.
//
// Holds the handle cruxd assigned at load time and sends every operation to
// cruxd under that handle.
type container struct {
	client *client           // Client the container was loaded through.
	id     agent.ContainerID // Handle assigned by cruxd.
}

// Starts the container's main process.
func (c *container) Start(ctx context.Context) error {
	return c.client.call(ctx, agent.EndpointStart, agent.MediaTypeStartRequest, agent.StartRequest{ID: c.id}, agent.MediaTypeStartResponse, &agent.StartResponse{})
}

// Stops the container's main process.
func (c *container) Stop(ctx context.Context) error {
	return c.client.call(ctx, agent.EndpointStop, agent.MediaTypeStopRequest, agent.StopRequest{ID: c.id}, agent.MediaTypeStopResponse, &agent.StopResponse{})
}

// Runs a command in the container and returns its exit code.
//
// Output messages are written to stdout or stderr as they arrive. The stream
// must end with an exit or error message; a stream that ends without one is
// an error.
func (c *container) Exec(ctx context.Context, stdout, stderr io.Writer, command string, opts agent.Options) (int, error) {
	req := agent.ExecRequest{ID: c.id, Command: command, Options: opts}
	resp, err := c.client.send(ctx, agent.EndpointExec, agent.MediaTypeExecRequest, req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if err := checkResponse(resp, agent.MediaTypeExecResponse.String()); err != nil {
		return 0, err
	}

	stream := agent.NewExecReader(resp.Body)
	for {
		msg, err := stream.Next()
		if errors.Is(err, io.EOF) {
			return 0, crex.Newf(ErrCruxd, "exec stream ended without a result")
		}
		if err != nil {
			return 0, crex.Wrap(ErrCruxd, err)
		}
		switch msg.Type {
		case agent.ExecOutput:
			if err := writeOutput(stdout, stderr, msg); err != nil {
				return 0, err
			}
		case agent.ExecExit:
			return msg.Code, nil
		case agent.ExecError:
			if msg.Error == nil {
				return 0, crex.Newf(ErrCruxd, "exec error message without an error")
			}
			return 0, crex.Wrap(ErrCruxd, msg.Error)
		default:
			return 0, crex.Newf(ErrCruxd, "unknown exec message type %q", msg.Type)
		}
	}
}

// Applies an uncompressed tar archive to the container's writable snapshot.
func (c *container) Copy(ctx context.Context, r io.Reader) error {
	resp, err := c.client.sendData(ctx, agent.EndpointCopy, agent.MediaTypeCopyRequest, agent.CopyRequest{ID: c.id}, agent.MediaTypeLayer, r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, agent.MediaTypeCopyResponse, &agent.CopyResponse{})
}

// Stores cfg as the pending configuration to be applied at commit.
func (c *container) Configure(ctx context.Context, cfg agent.ImageConfig) error {
	req := agent.ConfigureRequest{ID: c.id, Config: cfg}
	return c.client.call(ctx, agent.EndpointConfig, agent.MediaTypeConfigureRequest, req, agent.MediaTypeConfigureResponse, &agent.ConfigureResponse{})
}

// Returns the container's current OCI config.
func (c *container) Inspect(ctx context.Context) (agent.ImageConfig, error) {
	var out agent.InspectContainerResponse
	req := agent.InspectContainerRequest{ID: c.id}
	if err := c.client.call(ctx, agent.EndpointInspectContainer, agent.MediaTypeInspectContainerRequest, req, agent.MediaTypeInspectContainerResponse, &out); err != nil {
		return agent.ImageConfig{}, err
	}
	return out.Config, nil
}

// Creates a new image from the container's current state.
func (c *container) Commit(ctx context.Context, by string) (agent.ImageRef, error) {
	var out agent.CommitResponse
	req := agent.CommitRequest{ID: c.id, By: by}
	if err := c.client.call(ctx, agent.EndpointCommit, agent.MediaTypeCommitRequest, req, agent.MediaTypeCommitResponse, &out); err != nil {
		return "", err
	}
	return out.Ref, nil
}

// Releases the container on the machine.
func (c *container) Destroy(ctx context.Context) error {
	return c.client.call(ctx, agent.EndpointDestroy, agent.MediaTypeDestroyRequest, agent.DestroyRequest{ID: c.id}, agent.MediaTypeDestroyResponse, &agent.DestroyResponse{})
}

// Writes the data of an output message to the writer for its channel.
func writeOutput(stdout, stderr io.Writer, msg agent.ExecResponse) error {
	var w io.Writer
	switch msg.Channel {
	case agent.ChannelStdout:
		w = stdout
	case agent.ChannelStderr:
		w = stderr
	default:
		return crex.Newf(ErrCruxd, "unknown output channel %q", msg.Channel)
	}
	if _, err := w.Write(msg.Data); err != nil {
		return crex.Wrap(ErrCruxd, err)
	}
	return nil
}
