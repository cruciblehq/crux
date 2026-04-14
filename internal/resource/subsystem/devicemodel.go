package subsystem

import (
	"syscall"

	"github.com/cruciblehq/crex"
)

// Device node type.
type deviceNodeType string

const (
	deviceNodeChar       deviceNodeType = "c" // Character device.
	deviceNodeBlock      deviceNodeType = "b" // Block device.
	deviceNodeUnbuffered deviceNodeType = "u" // Unbuffered (character) device.
	deviceNodeFIFO       deviceNodeType = "p" // Named pipe (FIFO).
)

// Argument keys for device provisioning grants.
type deviceArg string

const (
	deviceArgPath     deviceArg = "path"      // Mount path inside the container.
	deviceArgType     deviceArg = "type"      // Device type (c/b/u/p).
	deviceArgMajor    deviceArg = "major"     // Major device number.
	deviceArgMinor    deviceArg = "minor"     // Minor device number.
	deviceArgFileMode deviceArg = "file_mode" // File permissions (decimal).
	deviceArgUID      deviceArg = "uid"       // Owner UID.
	deviceArgGID      deviceArg = "gid"       // Owner GID.
)

// Declares a device node to provision inside the container.
//
// Makes a character or block device available in the container's filesystem.
// The runtime creates the node via mknod(2) at the given path. This is purely
// provisioning: the cgroup device permission that allows processes to
// open the device lives in [CgroupSubsystem]. The zero value provisions nothing.
type device struct {
	Name     string         `codec:"name"`                        // Device name (e.g., "null", "zero", "urandom", "fuse", "nvidia0").
	Path     string         `codec:"path,omitempty"`              // Mount path inside the container.
	Type     deviceNodeType `codec:"type,omitempty"`              // Device type: "c" (char), "b" (block), "u" (unbuffered), "p" (FIFO).
	Major    uint32         `codec:"major,omitempty"`             // Major number identifying the kernel driver.
	Minor    uint32         `codec:"minor,omitempty"`             // Minor number identifying the specific device instance.
	FileMode uint16         `codec:"file_mode,omitempty"`         // File permissions (decimal). Zero means no permissions (mode 0000).
	UID      uint32         `codec:"uid,omitempty,default=65534"` // Device owner UID. Default is nobody (65534).
	GID      uint32         `codec:"gid,omitempty,default=65534"` // Device owner GID. Default is nogroup (65534).
}

// Merges another device's file mode into this one.
//
// Identity fields (name, path, type, major, minor, uid, gid) must match
// exactly or the merge returns a conflict. The upper 3 mode bits (setuid,
// setgid, and sticky) are also identity: if they differ, it is a conflict.
// The lower 9 permission bits are relaxed via bitwise OR. Returns true if
// the merge changed the effective file mode.
func (d *device) merge(other *device) (bool, error) {
	if d.Path != other.Path ||
		d.Type != other.Type ||
		d.Major != other.Major ||
		d.Minor != other.Minor ||
		d.UID != other.UID ||
		d.GID != other.GID {
		return false, crex.Wrapf(ErrGrantConflict, "device %s already provisioned with different properties", d.Name)
	}

	const specialBits = syscall.S_ISUID | syscall.S_ISGID | syscall.S_ISVTX
	if d.FileMode&specialBits != other.FileMode&specialBits {
		return false, crex.Wrapf(ErrGrantConflict, "device %s: conflicting special mode bits", d.Name)
	}

	merged := d.FileMode | other.FileMode
	if merged == d.FileMode {
		return false, nil
	}
	d.FileMode = merged
	return true, nil
}
