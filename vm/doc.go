// Package vm manages the lifecycle of a Linux virtual machine for running
// containers on non-Linux hosts.
//
// On macOS, Crucible needs a Linux VM to run OCI container images. This
// package wraps Lima to provide VM creation, startup, shutdown, and command
// execution. The VM runs containerd as its container runtime. On Linux,
// containers run natively and this package is not used. Attempting to call any
// function on a non-darwin platform returns [ErrUnsupportedPlatform].
package vm
