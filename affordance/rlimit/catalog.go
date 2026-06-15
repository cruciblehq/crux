package rlimit

// Prefix the kernel uses for the RLIMIT_<UPPER> constants. AGL grants use the
// lowercase short form (e.g. "nofile") which is upper-cased and prefixed.
const rlimitNamePrefix = "RLIMIT_"

// Keyword accepted in place of an integer limit to request no bound, mapping
// to RLIM_INFINITY.
const limitUnlimited = "unlimited"

// Positions of the positional arguments in an rlimit grant.
const (
	argResource = 0 // Resource short name.
	argSoft     = 1 // Soft limit value.
	argHard     = 2 // Hard limit value.
)

// Labels used in diagnostics for the two limit values.
const (
	labelSoft = "soft" // Soft limit label.
	labelHard = "hard" // Hard limit label.
)

// Numeric base and bit size used when parsing limit integers.
const (
	limitBase = 0  // Accept decimal, octal, and hexadecimal integer literals.
	limitBits = 64 // Limits are unsigned 64-bit values.
)

// All known POSIX resource short names accepted in .rlimit grants.
//
// The kernel exposes these via the RLIMIT_<UPPER> constants. AGL grants use
// the lowercase short form (e.g. "nofile" rather than "RLIMIT_NOFILE").
var knownResources = map[string]struct{}{
	"as":         {},
	"core":       {},
	"cpu":        {},
	"data":       {},
	"fsize":      {},
	"locks":      {},
	"memlock":    {},
	"msgqueue":   {},
	"nice":       {},
	"nofile":     {},
	"nproc":      {},
	"rss":        {},
	"rtprio":     {},
	"rttime":     {},
	"sigpending": {},
	"stack":      {},
}
