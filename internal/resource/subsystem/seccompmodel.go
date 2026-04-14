package subsystem

import (
	"regexp"
	"strings"

	"github.com/cruciblehq/crex"
)

// Matches a valid syscall name (lowercase letters, digits, and underscores).
var validSyscallName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Allows one syscall, optionally constrained by argument filters.
//
// When Args is empty, every invocation of the named syscall is permitted.
// When Args is non-empty, all conditions must match (AND) for the syscall
// to be allowed; invocations that don't match the conditions are killed.
// Multiple rules for the same syscall with different Args are OR'd by the
// BPF compiler: the syscall is allowed if any rule matches.
type seccomp struct {
	Syscall string       `codec:"syscall"`        // Syscall name (e.g., "socket", "read", "openat").
	Args    []seccompArg `codec:"args,omitempty"` // Argument conditions. All must match (AND).
}

// Constrains one syscall argument by position.
//
// Arg selects which argument (0-5). The comparison is Op(arg, Val). For
// seccompOpMaskedEq the comparison is (arg & Mask) == Val.
type seccompArg struct {
	Arg  uint8     `codec:"arg"`            // Argument position (0-5).
	Op   seccompOp `codec:"op"`             // Comparison operator.
	Val  uint64    `codec:"val"`            // Comparison value. For MaskedEq, the expected result.
	Mask uint64    `codec:"mask,omitempty"` // Bitmask. Only used with MaskedEq: (arg & Mask) == Val.
}

// Parses a compact expression into a seccomp rule.
//
// The string form is a bare syscall name (e.g., "read", "socket"). Argument
// filters are specified via the structured sub-args form only.
func (r *seccomp) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		return crex.Wrapf(ErrSeccompExpression, "expression is empty")
	}
	if !validSyscallName.MatchString(s) {
		return crex.Wrapf(ErrSeccompExpression, "invalid syscall name %q", s)
	}
	r.Syscall = s
	r.Args = nil
	return nil
}
