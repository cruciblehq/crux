// Package volume implements the .volume AGL subsystem.
//
// A .volume grant declares persistent storage at a given container path. The
// platform provisions and manages the underlying storage; the container sees
// it as a regular mount point. On the local provider this is a directory inside
// the VM. On cloud providers it is a managed persistent disk. Volumes survive
// container restarts and are not tied to the container lifecycle.
//
// Grant syntax:
//
//	.volume <destination> [r|rw]
//
// destination is the absolute path inside the container where the volume is
// mounted. The optional access token is "r" for a read-only mount or "rw" for
// a read-write mount; it defaults to "r" when omitted, keeping the baseline
// maximally restrictive.
package volume
