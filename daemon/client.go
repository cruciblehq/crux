package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"net"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/paths"
	"github.com/cruciblehq/spec/protocol"
)

// Client for communicating with the cruxd daemon.
//
// Each method opens a connection, sends a command, reads the response,
// and closes the connection. The daemon expects newline-delimited JSON
// over a Unix domain socket.
type Client struct {
	socketPath string
}

// Creates a new daemon client using the default socket path.
func NewClient() *Client {
	return &Client{socketPath: paths.DaemonSocket()}
}

// Creates a new daemon client targeting the given socket path.
func NewClientWithSocket(socketPath string) *Client {
	return &Client{socketPath: socketPath}
}

// Sends a build request to the daemon and waits for the result.
func (c *Client) Build(ctx context.Context, req *protocol.BuildRequest) (*protocol.BuildResult, error) {
	data, err := c.send(ctx, protocol.CmdBuild, req)
	if err != nil {
		return nil, err
	}

	result, err := protocol.DecodePayload[protocol.BuildResult](data)
	if err != nil {
		return nil, crex.Wrap(ErrRequest, err)
	}

	return result, nil
}

// Sends a status request to the daemon.
func (c *Client) Status(ctx context.Context) (*protocol.StatusResult, error) {
	data, err := c.send(ctx, protocol.CmdStatus, nil)
	if err != nil {
		return nil, err
	}

	result, err := protocol.DecodePayload[protocol.StatusResult](data)
	if err != nil {
		return nil, crex.Wrap(ErrRequest, err)
	}

	return result, nil
}

// Sends a shutdown request to the daemon.
func (c *Client) Shutdown(ctx context.Context) error {
	_, err := c.send(ctx, protocol.CmdShutdown, nil)
	return err
}

// Sends an image-import request to the daemon.
func (c *Client) ImageImport(ctx context.Context, req *protocol.ImageImportRequest) error {
	_, err := c.send(ctx, protocol.CmdImageImport, req)
	return err
}

// Sends an image-start request to the daemon.
func (c *Client) ImageStart(ctx context.Context, req *protocol.ImageStartRequest) error {
	_, err := c.send(ctx, protocol.CmdImageStart, req)
	return err
}

// Sends an image-destroy request to the daemon.
func (c *Client) ImageDestroy(ctx context.Context, req *protocol.ImageDestroyRequest) error {
	_, err := c.send(ctx, protocol.CmdImageDestroy, req)
	return err
}

// Sends a container-stop request to the daemon.
func (c *Client) ContainerStop(ctx context.Context, req *protocol.ContainerStopRequest) error {
	_, err := c.send(ctx, protocol.CmdContainerStop, req)
	return err
}

// Sends a container-destroy request to the daemon.
func (c *Client) ContainerDestroy(ctx context.Context, req *protocol.ContainerDestroyRequest) error {
	_, err := c.send(ctx, protocol.CmdContainerDestroy, req)
	return err
}

// Sends a container-status request to the daemon and returns the result.
func (c *Client) ContainerStatus(ctx context.Context, req *protocol.ContainerStatusRequest) (*protocol.ContainerStatusResult, error) {
	data, err := c.send(ctx, protocol.CmdContainerStatus, req)
	if err != nil {
		return nil, err
	}

	result, err := protocol.DecodePayload[protocol.ContainerStatusResult](data)
	if err != nil {
		return nil, crex.Wrap(ErrRequest, err)
	}

	return result, nil
}

// Sends a container-exec request to the daemon and returns the result.
func (c *Client) ContainerExec(ctx context.Context, req *protocol.ContainerExecRequest) (*protocol.ContainerExecResult, error) {
	data, err := c.send(ctx, protocol.CmdContainerExec, req)
	if err != nil {
		return nil, err
	}

	result, err := protocol.DecodePayload[protocol.ContainerExecResult](data)
	if err != nil {
		return nil, crex.Wrap(ErrRequest, err)
	}

	return result, nil
}

// Sends a container-update request to the daemon.
func (c *Client) ContainerUpdate(ctx context.Context, req *protocol.ContainerUpdateRequest) error {
	_, err := c.send(ctx, protocol.CmdContainerUpdate, req)
	return err
}

// Sends a command and returns the raw response payload.
//
// Opens a connection, writes the encoded envelope followed by a newline,
// reads the response line, and decodes it. Returns an error if the
// daemon responds with [protocol.CmdError].
func (c *Client) send(ctx context.Context, cmd protocol.Command, payload any) (json.RawMessage, error) {
	conn, err := c.dial(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	data, err := protocol.Encode(cmd, payload)
	if err != nil {
		return nil, crex.Wrap(ErrRequest, err)
	}
	data = append(data, byte(10))

	if _, err := conn.Write(data); err != nil {
		return nil, crex.Wrap(ErrConnection, err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes(byte(10))
	if err != nil {
		return nil, crex.Wrap(ErrConnection, err)
	}

	env, respPayload, err := protocol.Decode(line)
	if err != nil {
		return nil, crex.Wrap(ErrRequest, err)
	}

	if env.Command == protocol.CmdError {
		errResult, err := protocol.DecodePayload[protocol.ErrorResult](respPayload)
		if err != nil {
			return nil, crex.Wrap(ErrRequest, err)
		}
		if errResult != nil {
			return nil, crex.Wrapf(ErrRequest, "%s", errResult.Message)
		}
		return nil, ErrRequest
	}

	return respPayload, nil
}

// Opens a Unix domain socket connection to the daemon.
func (c *Client) dial(ctx context.Context) (net.Conn, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", c.socketPath)
	if err != nil {
		return nil, crex.Wrap(ErrNotRunning, err)
	}
	return conn, nil
}
