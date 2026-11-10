package compute

import (
	"bufio"
	"context"
	"io"
	"net/url"

	speccompute "github.com/cruciblehq/spec/compute"
	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/crex"
)

// Make sure that the HTTP client satisfies the spec.
var _ speccompute.Container = (*cruxdContainer)(nil)

// Container on a machine, reached through cruxd over HTTP.
//
// Holds the handle cruxd assigned at load time and sends every operation to
// cruxd under that handle.
type cruxdContainer struct {
	runtime *cruxdRuntime           // Runtime the container was loaded through.
	id      speccompute.ContainerID // Handle assigned by cruxd.
}

// Starts the container's main process.
func (c *cruxdContainer) Start(ctx context.Context) error {
	return c.runtime.call(ctx, speccompute.EndpointStart, speccompute.MediaTypeStartRequest, speccompute.StartRequest{ID: c.id}, "", nil)
}

// Stops the container's main process.
func (c *cruxdContainer) Stop(ctx context.Context) error {
	return c.runtime.call(ctx, speccompute.EndpointStop, speccompute.MediaTypeStopRequest, speccompute.StopRequest{ID: c.id}, "", nil)
}

// Runs a command in the container and returns its exit code.
//
// Output messages are written to stdout or stderr as they arrive. The stream
// must end with an exit or error message; a stream that ends without one is
// an error.
func (c *cruxdContainer) Exec(ctx context.Context, stdout, stderr io.Writer, command string, opts speccompute.Options) (int, error) {
	req := speccompute.ExecRequest{ID: c.id, Command: command, Options: opts}
	resp, err := c.runtime.send(ctx, speccompute.EndpointExec, speccompute.MediaTypeExecRequest, req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if err := checkResponse(resp, speccompute.MediaTypeExecResponse.String()); err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var msg speccompute.ExecResponse
		if err := codec.Unmarshal(scanner.Bytes(), &msg, codec.JSON); err != nil {
			return 0, crex.Wrap(ErrCruxd, err)
		}
		switch msg.Type {
		case speccompute.ExecOutput:
			if err := writeOutput(stdout, stderr, msg); err != nil {
				return 0, err
			}
		case speccompute.ExecExit:
			return msg.Code, nil
		case speccompute.ExecError:
			if msg.Error == nil {
				return 0, crex.Newf(ErrCruxd, "exec error message without an error")
			}
			return 0, crex.Wrap(ErrCruxd, msg.Error)
		default:
			return 0, crex.Newf(ErrCruxd, "unknown exec message type %q", msg.Type)
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, crex.Wrap(ErrCruxd, err)
	}
	return 0, crex.Newf(ErrCruxd, "exec stream ended without a result")
}

// Applies an uncompressed tar archive to the container's writable snapshot.
func (c *cruxdContainer) Copy(ctx context.Context, r io.Reader) error {
	query := url.Values{speccompute.QueryParamID: {string(c.id)}}
	resp, err := c.runtime.post(ctx, speccompute.EndpointCopy, query, speccompute.MediaTypeLayer.String(), r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkResponse(resp, "")
}

// Stores cfg as the pending configuration to be applied at commit.
func (c *cruxdContainer) Configure(ctx context.Context, cfg speccompute.ImageConfig) error {
	req := speccompute.ConfigureRequest{ID: c.id, Config: cfg}
	return c.runtime.call(ctx, speccompute.EndpointConfig, speccompute.MediaTypeConfigureRequest, req, "", nil)
}

// Returns the container's current OCI config.
func (c *cruxdContainer) Inspect(ctx context.Context) (speccompute.ImageConfig, error) {
	var out speccompute.InspectContainerResponse
	req := speccompute.InspectContainerRequest{ID: c.id}
	if err := c.runtime.call(ctx, speccompute.EndpointInspectContainer, speccompute.MediaTypeInspectContainerRequest, req, speccompute.MediaTypeInspectContainerResponse, &out); err != nil {
		return speccompute.ImageConfig{}, err
	}
	return out.Config, nil
}

// Creates a new image from the container's current state.
func (c *cruxdContainer) Commit(ctx context.Context, by string) (speccompute.ImageRef, error) {
	var out speccompute.CommitResponse
	req := speccompute.CommitRequest{ID: c.id, By: by}
	if err := c.runtime.call(ctx, speccompute.EndpointCommit, speccompute.MediaTypeCommitRequest, req, speccompute.MediaTypeCommitResponse, &out); err != nil {
		return "", err
	}
	return out.Ref, nil
}

// Releases the container on the machine.
func (c *cruxdContainer) Destroy(ctx context.Context) error {
	return c.runtime.call(ctx, speccompute.EndpointDestroy, speccompute.MediaTypeDestroyRequest, speccompute.DestroyRequest{ID: c.id}, "", nil)
}

// Writes the data of an output message to the writer for its channel.
func writeOutput(stdout, stderr io.Writer, msg speccompute.ExecResponse) error {
	var w io.Writer
	switch msg.Channel {
	case speccompute.ChannelStdout:
		w = stdout
	case speccompute.ChannelStderr:
		w = stderr
	default:
		return crex.Newf(ErrCruxd, "unknown output channel %q", msg.Channel)
	}
	if _, err := w.Write(msg.Data); err != nil {
		return crex.Wrap(ErrCruxd, err)
	}
	return nil
}
