// Package daemon provides a client for communicating with the cruxd daemon
// over its Unix domain socket.
//
// The client sends protocol-encoded commands and receives responses. Each
// call opens a connection, writes the request, reads the response, and
// closes the connection. The daemon uses newline-delimited JSON as its
// wire format.
package daemon
