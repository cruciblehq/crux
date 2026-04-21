package fcap

import "github.com/cruciblehq/crex"

// Selects how file capabilities are granted on a binary.
//
// File capabilities are extended attributes on executables that the kernel
// evaluates during execve to compute the new process's capability sets.
// The mode determines whether capabilities become file-permitted (effective
// on exec) or file-inheritable (only effective if the caller also holds
// them in its inheritable set).
type Mode string

const (
	ModeEffective   Mode = "effective"   // File-permitted + effective bit. Caps are immediately effective after exec.
	ModeInheritable Mode = "inheritable" // File-inheritable. Caps only effective if caller holds them in inheritable set.
)

// Parses a mode string, returning an error for unknown values.
func ParseMode(s string) (Mode, error) {
	m := Mode(s)
	switch m {
	case ModeEffective, ModeInheritable:
		return m, nil
	default:
		return "", crex.Wrapf(ErrInvalidRule, "unknown fcap mode %q", s)
	}
}

// Rule expression for file capability grants.
//
// Each rule names a binary path, a mode selecting which file capability
// sets to populate, and the capability names.
type Grant struct {
	Mode Mode     `codec:"mode"` // Which file capability sets to populate.
	Path string   `codec:"path"` // Absolute path to the executable.
	Caps []string `codec:"caps"` // Capabilities to grant, without CAP_ prefix.
}
