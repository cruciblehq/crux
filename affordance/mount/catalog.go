package mount

// In-kernel filesystem types accepted as the first positional argument of a
// .mount grant.
const (
	fsTmpfs   = "tmpfs"   // Temporary in-memory filesystem.
	fsProc    = "proc"    // Process information pseudo-filesystem.
	fsSysfs   = "sysfs"   // Kernel object pseudo-filesystem.
	fsDevpts  = "devpts"  // Pseudo-terminal slave filesystem.
	fsMqueue  = "mqueue"  // POSIX message queue filesystem.
	fsCgroup2 = "cgroup2" // Unified cgroup v2 hierarchy.
)

// Keyword arguments accepted on tmpfs mounts.
const (
	kwSize = "size" // Upper bound on the filesystem's memory usage.
	kwMode = "mode" // Root directory permission bits, in octal.
)

// Mount options applied to every mount as a security baseline.
const (
	optNosuid = "nosuid" // Ignore set-user-ID and set-group-ID bits.
	optNodev  = "nodev"  // Disallow access to device special files.
	optNoexec = "noexec" // Disallow program execution.
)

// Positions of the positional arguments in a mount grant.
const (
	argType = 0 // Filesystem type token.
	argDest = 1 // Absolute destination path.
)

// Number of positional arguments a mount grant requires.
const mountArgCount = 2

// Root directory permission bits applied to tmpfs mounts unless overridden by
// the mode keyword argument.
const defaultTmpfsMode = "1777"

// In-kernel filesystem types accepted in a mount grant.
//
// Only types that are provided by the kernel itself are allowed; bind mounts
// from host paths are deliberately excluded to keep the security boundary
// between host and container intact.
var knownTypes = map[string]struct{}{
	fsTmpfs:   {},
	fsProc:    {},
	fsSysfs:   {},
	fsDevpts:  {},
	fsMqueue:  {},
	fsCgroup2: {},
}
