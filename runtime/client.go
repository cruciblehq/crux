//go:build !darwin && !linux

package runtime

import containerd "github.com/containerd/containerd/v2/client"

// Returns [ErrUnsupportedPlatform].
func newContainerdClient(_ string) (*containerd.Client, error) {
	return nil, ErrUnsupportedPlatform
}
