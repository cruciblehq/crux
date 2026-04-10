package subsystem

import (
	"context"
	"strings"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
)

// Cap permission verb constants.
const (
	CapVerbEffective   = "effective"   // Effective + permitted + bounding (effective immediately, survives exec, does not auto-inherit).
	CapVerbInheritable = "inheritable" // Permitted + inheritable + ambient + bounding (auto-inherits across exec via ambient).
	CapVerbPermitted   = "permitted"   // Permitted + bounding (raisable on demand, not effective by default).
	CapVerbBound       = "bound"       // Bounding only (exec ceiling for child processes).
)

// Controls which privileged kernel operations are allowed to the container.
//
// Linux capabilities divide root privilege into individual units (e.g., CHOWN,
// NET_BIND_SERVICE, SYS_PTRACE). Each field holds the names of capabilities
// granted to one of the five kernel sets: effective, permitted, inheritable,
// bounding, and ambient. The zero value grants no capabilities. Use the Grant
// methods to mutate; they enforce the kernel invariants between sets. When no
// verb is specified in the expression (just the capability name), all five
// sets are populated.
type Caps struct {
	Effective   []string `codec:"effective,omitempty"`   // Effective capability set.
	Permitted   []string `codec:"permitted,omitempty"`   // Permitted capability set.
	Inheritable []string `codec:"inheritable,omitempty"` // Inheritable capability set.
	Bounding    []string `codec:"bounding,omitempty"`    // Bounding capability set.
	Ambient     []string `codec:"ambient,omitempty"`     // Ambient capability set.
}

// Grants a capability to all five sets.
//
// This is the broadest grant: the capability is effective immediately,
// survives exec, and auto-inherits to child processes. This is the default
// when no verb is specified in the expression.
func (c *Caps) Grant(cap string) {
	appendUnique(&c.Effective, cap)
	appendUnique(&c.Permitted, cap)
	appendUnique(&c.Inheritable, cap)
	appendUnique(&c.Bounding, cap)
	appendUnique(&c.Ambient, cap)
}

// Grants a capability to the effective, permitted, and bounding sets.
//
// The capability is effective immediately and survives exec (via bounding),
// but does not auto-inherit to child processes. Useful for capabilities the
// service itself needs.
func (c *Caps) GrantEffective(cap string) {
	appendUnique(&c.Effective, cap)
	appendUnique(&c.Permitted, cap)
	appendUnique(&c.Bounding, cap)
}

// Grants a capability that auto-inherits across exec.
//
// The capability is not effective in the current process, but after execve
// the ambient set automatically raises it into the child's effective and
// permitted sets. Useful for capabilities a service's children need but the
// parent doesn't use directly.
func (c *Caps) GrantInheritable(cap string) {
	appendUnique(&c.Permitted, cap)
	appendUnique(&c.Inheritable, cap)
	appendUnique(&c.Ambient, cap)
	appendUnique(&c.Bounding, cap)
}

// Grants a capability to the permitted and bounding sets.
//
// The process may raise it into its effective set at will, and the bounding
// set allows it to persist across exec. Not effective by default, and does
// not auto-inherit. Useful for capabilities that are only needed for specific
// operations.
func (c *Caps) GrantPermitted(cap string) {
	appendUnique(&c.Permitted, cap)
	appendUnique(&c.Bounding, cap)
}

// Grants a capability only in the bounding set.
//
// This acts as an exec ceiling: child processes may receive this capability
// (via file caps or ambient), but the current process cannot use it. Useful
// for capabilities that are only needed by child processes.
func (c *Caps) GrantBound(cap string) {
	appendUnique(&c.Bounding, cap)
}

// Implements [Subsystem] for [manifest.DomainCap].
type CapsSubsystem struct{}

// Builds a caps grant by validating the expression "[verb] <name>".
//
// When no verb is given (single field), all five sets are populated. When a
// verb is present: effective, inheritable, permitted, bound. Returns a single
// grant with the validated expression.
func (s *CapsSubsystem) Build(_ context.Context, domain Domain, input manifest.Grant) ([]manifest.Grant, error) {
	crex.Assertf(domain == DomainCap, "unexpected domain %q", domain)
	_, err := parseCaps(input.Expr)
	if err != nil {
		return nil, err
	}
	return []manifest.Grant{{Subsystem: string(domain), Expr: input.Expr}}, nil
}

// Applies caps grants to a runtime.
func (s *CapsSubsystem) Apply(_ context.Context, _ Domain, _ manifest.Grant) error {
	crex.Assert(false, "not implemented")
	return nil
}

// Parses a caps expression "[verb] <name>" into a [Caps] config.
func parseCaps(expr string) (Caps, error) {
	fields := strings.Fields(expr)

	var verb, name string
	switch len(fields) {
	case 1:
		name = fields[0]
	case 2:
		verb, name = fields[0], fields[1]
	default:
		return Caps{}, crex.Wrapf(ErrSandboxExpression, "invalid expression %q", expr)
	}

	var c Caps
	switch verb {
	case "":
		c.Grant(name)
	case CapVerbEffective:
		c.GrantEffective(name)
	case CapVerbInheritable:
		c.GrantInheritable(name)
	case CapVerbPermitted:
		c.GrantPermitted(name)
	case CapVerbBound:
		c.GrantBound(name)
	default:
		return Caps{}, crex.Wrapf(ErrSandboxExpression, "unknown verb %q", verb)
	}
	return c, nil
}
