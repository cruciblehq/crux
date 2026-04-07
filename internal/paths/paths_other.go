//go:build !darwin && !linux

package paths

// Path to the containerd Unix socket for an instance.
//
// Not supported on this platform.
func ContainerdSocket(_ string) string {
	panic("containerd socket path is not available on this platform")
}
