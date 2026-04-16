package subsystem

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
