//go:build !windows

package runtime

const (

	// Socket filename for the containerd gRPC endpoint.
	containerdSock = "containerd.sock"
)
