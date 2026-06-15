// Package device implements the .device AGL subsystem.
//
// A .device grant provisions a device node in the container's OCI runtime
// spec. It creates the node only; access to the device is gated separately by
// the cgroup device controller through the .cgroup subsystem. The two are
// distinct kernel mechanisms and a usable device requires both: the node must
// exist in the container and the cgroup filter must permit its major:minor.
//
// Grant syntax:
//
//	.device <type> <path> <major> <minor> [mode=<octal>] [uid=<n>] [gid=<n>]
//
// type is the device type: c or u for character devices, b for a block
// device, or p for a FIFO. path is the absolute node path inside the
// container. major and minor are the device numbers. mode sets the node
// permission bits in octal, and uid and gid set its owner; all three are
// optional.
package device
