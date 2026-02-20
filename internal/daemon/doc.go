// Package daemon provides a client for communicating with the cruxd daemon
// over its Unix domain socket.
//
// The client sends protocol-encoded commands and receives responses. Each
// call opens a connection, writes the request, reads the response, and
// closes the connection. The daemon uses newline-delimited JSON as its
// wire format.
//
// Checking daemon status:
//
//	client := daemon.NewClient()
//	status, err := client.Status(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Sending a build request:
//
//	client := daemon.NewClient()
//	result, err := client.Build(ctx, &protocol.BuildRequest{
//	    Dir: "/path/to/project",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
package daemon
