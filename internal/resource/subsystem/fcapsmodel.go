package subsystem

// Declares file capabilities on a binary inside the container.
//
// Controls which capabilities a specific executable receives on exec. File
// capabilities are extended attributes (security.capability) on binaries that
// the kernel evaluates during execve to compute the new process's capability
// sets. For example, /usr/bin/ping can hold CAP_NET_RAW as a file-permitted
// cap so unprivileged users can send ICMP packets. Each listed capability
// must also be in the bounding set (via GrantBound or broader) for the file
// cap to take effect at exec time.
type fcap struct {
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
func (f *fcap) grantEffective(cap string) {
	appendUnique(&f.Permitted, cap)
	f.Effective = true
}

// Grants a file-inheritable capability.
//
// The capability only takes effect if the calling process also holds it in
// its inheritable set (see caps.GrantHeritable). This is useful for
// capabilities that should propagate through a chain of execs only when
// the parent explicitly opts in.
func (f *fcap) grantInheritable(cap string) {
	appendUnique(&f.Inheritable, cap)
}

// Applies a file capability verb to this model for a single capability name.
func (f *fcap) grant(verb fcapVerb, name string) {
	switch verb {
	case fcapVerbEffective:
		f.grantEffective(name)
	case fcapVerbInheritable:
		f.grantInheritable(name)
	}
}
