//go:build linux

package paths

// Path to the containerd Unix socket.
//
// On Linux containerd runs as a system service at a fixed path.
//
//	/run/containerd/containerd.sock
func ContainerdSocket(_ string) string {
	return "/run/containerd/containerd.sock"
}
