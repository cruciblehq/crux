//go:build darwin || linux

package local

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"syscall"
	"time"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/compute/internal/provider"
	"github.com/cruciblehq/crux/internal/paths"
	"github.com/cruciblehq/spec/protocol"
)

// Communicates with cruxd over a Unix domain socket.
type client struct {
	name string
}

// Returns a new [provider.Client] for the given instance.
func newClient(name string) (provider.Client, error) {
	return &client{name: name}, nil
}

// Sends a build request and waits for the result.
func (c *client) Build(ctx context.Context, req *protocol.BuildRequest) (*protocol.BuildResult, error) {
	data, err := c.send(ctx, protocol.CmdBuild, req)
	if err != nil {
		return nil, err
	}
	result, err := protocol.DecodePayload[protocol.BuildResult](data)
	if err != nil {
		return nil, crex.Wrap(provider.ErrRequestFailed, err)
	}
	return result, nil
}

// Returns the instance's current status.
func (c *client) Status(ctx context.Context) (*protocol.StatusResult, error) {
	data, err := c.send(ctx, protocol.CmdStatus, nil)
	if err != nil {
		return nil, err
	}
	result, err := protocol.DecodePayload[protocol.StatusResult](data)
	if err != nil {
		return nil, crex.Wrap(provider.ErrRequestFailed, err)
	}
	return result, nil
}

// Imports a container image from a local archive.
func (c *client) ImageImport(ctx context.Context, req *protocol.ImageImportRequest) error {
	_, err := c.send(ctx, protocol.CmdImageImport, req)
	return err
}

// Starts a container from a previously imported image.
func (c *client) ImageStart(ctx context.Context, req *protocol.ImageStartRequest) error {
	_, err := c.send(ctx, protocol.CmdImageStart, req)
	return err
}

// Removes a container image and its associated resources.
func (c *client) ImageDestroy(ctx context.Context, req *protocol.ImageDestroyRequest) error {
	_, err := c.send(ctx, protocol.CmdImageDestroy, req)
	return err
}

// Stops a running container.
func (c *client) ContainerStop(ctx context.Context, req *protocol.ContainerStopRequest) error {
	_, err := c.send(ctx, protocol.CmdContainerStop, req)
	return err
}

// Destroys a container and its filesystem state.
func (c *client) ContainerDestroy(ctx context.Context, req *protocol.ContainerDestroyRequest) error {
	_, err := c.send(ctx, protocol.CmdContainerDestroy, req)
	return err
}

// Returns the current state of a container.
func (c *client) ContainerStatus(ctx context.Context, req *protocol.ContainerStatusRequest) (*protocol.ContainerStatusResult, error) {
	data, err := c.send(ctx, protocol.CmdContainerStatus, req)
	if err != nil {
		return nil, err
	}
	result, err := protocol.DecodePayload[protocol.ContainerStatusResult](data)
	if err != nil {
		return nil, crex.Wrap(provider.ErrRequestFailed, err)
	}
	return result, nil
}

// Executes a command inside a running container.
func (c *client) ContainerExec(ctx context.Context, req *protocol.ContainerExecRequest) (*protocol.ContainerExecResult, error) {
	data, err := c.send(ctx, protocol.CmdContainerExec, req)
	if err != nil {
		return nil, err
	}
	result, err := protocol.DecodePayload[protocol.ContainerExecResult](data)
	if err != nil {
		return nil, crex.Wrap(provider.ErrRequestFailed, err)
	}
	return result, nil
}

// Updates a running container's configuration.
func (c *client) ContainerUpdate(ctx context.Context, req *protocol.ContainerUpdateRequest) error {
	_, err := c.send(ctx, protocol.CmdContainerUpdate, req)
	return err
}

// Sends a command and returns the raw response payload.
func (c *client) send(ctx context.Context, cmd protocol.Command, payload any) (json.RawMessage, error) {
	conn, err := dial(ctx, c.name)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Unblock the socket read when the context is cancelled (e.g. SIGINT) by
	// setting a deadline in the past, causing ReadBytes to return immediately
	// with a timeout error.
	go func() {
		<-ctx.Done()
		conn.SetReadDeadline(time.Now())
	}()

	if err := writeCommand(conn, cmd, payload); err != nil {
		return nil, err
	}

	env, data, err := readResponse(conn)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	if err := checkErrorResponse(env, data); err != nil {
		return nil, err
	}

	return data, nil
}

// Encodes a command with its payload and writes it to the connection.
func writeCommand(conn net.Conn, cmd protocol.Command, payload any) error {
	data, err := protocol.Encode(cmd, payload)
	if err != nil {
		return crex.Wrap(provider.ErrRequestFailed, err)
	}
	data = append(data, '\n')

	if _, err := conn.Write(data); err != nil {
		return crex.Wrap(provider.ErrConnectionFailed, err)
	}
	return nil
}

// Reads a newline-delimited response and decodes the envelope.
func readResponse(conn net.Conn) (*protocol.Envelope, json.RawMessage, error) {
	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, nil, crex.Wrap(provider.ErrConnectionFailed, err)
	}

	env, payload, err := protocol.Decode(line)
	if err != nil {
		return nil, nil, crex.Wrap(provider.ErrRequestFailed, err)
	}
	return env, payload, nil
}

// Returns an error if the envelope carries an error response.
func checkErrorResponse(env *protocol.Envelope, data json.RawMessage) error {
	if env.Command != protocol.CmdError {
		return nil
	}

	errResult, err := protocol.DecodePayload[protocol.ErrorResult](data)
	if err != nil {
		return crex.Wrap(provider.ErrRequestFailed, err)
	}
	if errResult != nil {
		return crex.Wrapf(provider.ErrRequestFailed, "%s", errResult.Message)
	}
	return provider.ErrRequestFailed
}

// Opens a Unix domain socket connection to a cruxd instance.
func dial(ctx context.Context, name string) (net.Conn, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", paths.CruxdSocket(name))
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			return nil, crex.Wrap(provider.ErrConnectionRefused, err)
		}
		if errors.Is(err, syscall.ENOENT) {
			return nil, crex.Wrap(provider.ErrNotRunning, err)
		}
		return nil, crex.Wrap(provider.ErrConnectionFailed, err)
	}
	return conn, nil
}
