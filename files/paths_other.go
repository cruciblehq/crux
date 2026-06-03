//go:build !darwin && !linux

package files

// Path to the containerd Unix socket for an instance.
//
// Not supported on this platform.
func ContainerdSocket(_ string) string {
	panic("containerd socket is not supported on this platform")
}
