package device

// Device node types that mirror the type field of an OCI device node.
const (
	typeChar       = "c" // Character device.
	typeBlock      = "b" // Block device.
	typeUnbuffered = "u" // Unbuffered character device.
	typeFIFO       = "p" // FIFO (named pipe).
)

// Keyword arguments accepted after the positional arguments.
const (
	kwMode = "mode" // Node permission bits, in octal.
	kwUID  = "uid"  // Owner user id.
	kwGID  = "gid"  // Owner group id.
)

// Positions of the required positional arguments in a device grant.
const (
	argType  = 0 // Device type token.
	argPath  = 1 // Absolute node path.
	argMajor = 2 // Device major number.
	argMinor = 3 // Device minor number.
)

// Number of positional arguments a device grant requires.
const deviceArgCount = 4

// Numeric bases and bit sizes used when parsing device grant values.
const (
	decimalBase   = 10 // Base for major, minor, uid, and gid values.
	octalBase     = 8  // Base for the mode value.
	deviceNumBits = 64 // Bit size for the major and minor numbers.
	idBits        = 32 // Bit size for the mode, uid, and gid values.
)

// Accepted device node types.
//
// "c" and "u" are character devices ("u" being unbuffered), "b" is a block
// device, and "p" is a FIFO.
var knownTypes = map[string]struct{}{
	typeChar:       {},
	typeBlock:      {},
	typeUnbuffered: {},
	typeFIFO:       {},
}
