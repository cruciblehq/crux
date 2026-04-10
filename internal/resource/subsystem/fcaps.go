package subsystem

import (
	"context"
	"strings"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
)

// Declares file capabilities on a binary inside the container.
//
// Controls which capabilities a specific executable receives on exec. File
// capabilities are extended attributes (security.capability) on binaries that
// the kernel evaluates during execve to compute the new process's capability
// sets. For example, /usr/bin/ping can hold CAP_NET_RAW as a file-permitted
// cap so unprivileged users can send ICMP packets. Each listed capability
// must also be in the bounding set (via GrantBound or broader) for the file
// cap to take effect at exec time.
type Fcap struct {
	Path        string   `codec:"path"`                  // Binary path inside the container (e.g., "/usr/bin/ping").
	Permitted   []string `codec:"permitted,omitempty"`   // File-permitted capability names.
	Inheritable []string `codec:"inheritable,omitempty"` // File-inheritable capability names.
	Effective   bool     `codec:"effective,omitempty"`   // If true, all new-permitted caps become effective on exec.
}

// Grants a file-permitted capability and sets the effective bit.
//
// After execve the capability is immediately effective in the new process.
// This is the common case for binaries that need a privilege unconditionally
// (e.g., ping needs NET_RAW).
func (f *Fcap) GrantEffective(cap string) {
	appendUnique(&f.Permitted, cap)
	f.Effective = true
}

// Grants a file-inheritable capability.
//
// The capability only takes effect if the calling process also holds it in
// its inheritable set (see Caps.GrantHeritable). This is useful for
// capabilities that should propagate through a chain of execs only when
// the parent explicitly opts in.
func (f *Fcap) GrantInheritable(cap string) {
	appendUnique(&f.Inheritable, cap)
}

// Implements [Subsystem] for [manifest.DomainFcap].
type FcapsSubsystem struct{}

// Builds an fcap grant by validating the expression "<verb> <path> <caps...>".
//
// Verbs: effective (file-permitted + effective bit), inheritable
// (file-inheritable only). Returns a single grant with the validated
// expression.
func (s *FcapsSubsystem) Build(_ context.Context, domain Domain, input manifest.Grant) ([]manifest.Grant, error) {
	crex.Assertf(domain == DomainFcap, "unexpected domain %q", domain)
	_, err := parseFcap(input.Expr)
	if err != nil {
		return nil, err
	}
	return []manifest.Grant{{Subsystem: string(domain), Expr: input.Expr}}, nil
}

// Applies fcap grants to a runtime.
func (s *FcapsSubsystem) Apply(_ context.Context, _ Domain, _ manifest.Grant) error {
	crex.Assert(false, "not implemented")
	return nil
}

// Parses an fcap expression "<verb> <path> <caps...>" into an [Fcap] config.
func parseFcap(expr string) (Fcap, error) {
	fields := strings.Fields(expr)
	if len(fields) < 3 {
		return Fcap{}, crex.Wrapf(ErrSandboxExpression, "invalid expression %q", expr)
	}
	verb, path, caps := fields[0], fields[1], fields[2:]

	var fc Fcap
	fc.Path = path

	switch verb {
	case "effective":
		for _, c := range caps {
			fc.GrantEffective(c)
		}
	case "inheritable":
		for _, c := range caps {
			fc.GrantInheritable(c)
		}
	default:
		return Fcap{}, crex.Wrapf(ErrSandboxExpression, "unknown verb %q", verb)
	}
	return fc, nil
}

// Appends s to the slice at dst if s is not already present.
func appendUnique(dst *[]string, s string) {
	for _, existing := range *dst {
		if existing == s {
			return
		}
	}
	*dst = append(*dst, s)
}
