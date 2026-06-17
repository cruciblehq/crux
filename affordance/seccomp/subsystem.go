package seccomp

import (
	"github.com/cruciblehq/crux/crex"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
)

// Implementation for seccomp.
//
// Holds a pointer to the OCI seccomp section of the unified spec, wired in
// at construction time. The shared baseline pre-populates the section with
// a deny-all default action and a single allow entry for exit_group; Build
// appends further allow entries.
type Subsystem struct {
	spec *specs.LinuxSeccomp
}

// Returns a Subsystem wired to mutate seccomp.
func New(spec *specs.LinuxSeccomp) *Subsystem {
	return &Subsystem{spec: spec}
}

// Returns the seccomp subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameSeccomp
}

// Applies a parsed grant to the wired-in section.
func (s *Subsystem) Build(g *agl.Model) error {
	if err := check(g); err != nil {
		return err
	}
	entries, err := parse(g)
	if err != nil {
		return err
	}
	for _, e := range entries {
		applyEntry(s.spec, e)
	}
	return nil
}

// Validates the grant's structural shape against what the seccomp subsystem accepts.
func check(g *agl.Model) error {
	if g.Where != nil {
		return crex.Newf(ErrInvalidGrant, "unexpected where clause in seccomp expression")
	}
	if len(g.Kwargs) != 0 {
		return crex.Newf(ErrInvalidGrant, "unexpected keyword arguments in seccomp expression")
	}
	if len(g.Args) == 0 {
		return crex.Newf(ErrInvalidGrant, "missing syscall name in seccomp expression")
	}
	if len(g.Args) > 2 {
		return crex.Newf(ErrInvalidGrant, "too many arguments in seccomp expression")
	}
	return nil
}

// Extracts and validates the grant's content into one or more syscall entries.
//
// The first argument is the syscall name, which may be followed by an optional
// sub-filter qualifier. The syscall name must be part of the x86_64 kernel ABI.
// Sub-filter qualifiers are only accepted for a curated set of syscalls, and
// must be one of a curated set of values. Unknown syscalls or qualifiers are
// errors.
func parse(g *agl.Model) ([]specs.LinuxSyscall, error) {
	knob := make([]string, 0, len(g.Args))
	for _, a := range g.Args {
		if a.Type != agl.ArgName {
			return nil, crex.Newf(ErrInvalidGrant, "expected name as argument in seccomp expression")
		}
		knob = append(knob, a.Value)
	}
	if _, ok := syscalls[knob[0]]; !ok {
		return nil, crex.Newf(ErrUnknownSyscall, "syscall %q is not part of the x86_64 kernel ABI", knob[0])
	}
	switch knob[0] {
	case sysIoctl:
		return expandSub(sysIoctl, 1, knob[1:], ioctlSubs)
	case sysFcntl:
		return expandSub(sysFcntl, 1, knob[1:], fcntlSubs)
	case sysPrctl:
		return expandSub(sysPrctl, 0, knob[1:], prctlSubs)
	default:
		if len(knob) > 1 {
			return nil, crex.Newf(ErrInvalidGrant, "unexpected sub-filter on syscall %q in seccomp expression", knob[0])
		}
		return []specs.LinuxSyscall{unconditionalAllow(knob[0])}, nil
	}
}

// Expands a curated sub-filter qualifier into one or more syscall entries.
//
// The sub-filter is looked up in subs; if found, an allow entry is produced for
// each value. An empty sub-filter produces a single unconditional allow entry.
// An unknown sub-filter is an error.
func expandSub(syscall string, argIndex uint, sub []string, subs map[string][]uint64) ([]specs.LinuxSyscall, error) {
	if len(sub) == 0 {
		return []specs.LinuxSyscall{unconditionalAllow(syscall)}, nil
	}
	vals, ok := subs[sub[0]]
	if !ok {
		return nil, crex.Newf(ErrInvalidGrant, "unknown %s sub-filter %q in seccomp expression", syscall, sub[0])
	}
	out := make([]specs.LinuxSyscall, 0, len(vals))
	for _, v := range vals {
		out = append(out, specs.LinuxSyscall{
			Names:  []string{syscall},
			Action: specs.ActAllow,
			Args: []specs.LinuxSeccompArg{{
				Index: argIndex,
				Value: v,
				Op:    specs.OpEqualTo,
			}},
		})
	}
	return out, nil
}

// Inserts entry into seccomp, deduping against existing entries.
//
// If entry is a single syscall allow rule, it replaces any prior entries for
// the same syscall. This dedup is intentionally self-sufficient and does not
// rely on any caller having already rejected duplicate grants.
func applyEntry(seccomp *specs.LinuxSeccomp, entry specs.LinuxSyscall) {
	if !isSingleNameAllow(entry) {
		seccomp.Syscalls = append(seccomp.Syscalls, cloneSyscall(entry))
		return
	}
	syscall := entry.Names[0]
	if len(entry.Args) == 0 {
		replaceWithUnconditional(seccomp, syscall, entry)
		return
	}
	if entryRedundantWith(seccomp, syscall, entry) {
		return
	}
	seccomp.Syscalls = append(seccomp.Syscalls, cloneSyscall(entry))
}

// Whether entry is a simple single-syscall allow rule.
func isSingleNameAllow(entry specs.LinuxSyscall) bool {
	return entry.Action == specs.ActAllow && len(entry.Names) == 1
}

// Replaces every prior allow entry for syscall with the unconditional one.
func replaceWithUnconditional(seccomp *specs.LinuxSeccomp, syscall string, entry specs.LinuxSyscall) {
	filtered := seccomp.Syscalls[:0]
	for _, existing := range seccomp.Syscalls {
		if isSingleNameAllow(existing) && existing.Names[0] == syscall {
			continue
		}
		filtered = append(filtered, existing)
	}
	seccomp.Syscalls = append(filtered, cloneSyscall(entry))
}

// Whether entry is already covered by an existing accumulated rule.
func entryRedundantWith(seccomp *specs.LinuxSeccomp, syscall string, entry specs.LinuxSyscall) bool {
	for _, existing := range seccomp.Syscalls {
		if syscallEqual(existing, entry) {
			return true
		}
	}
	for _, existing := range seccomp.Syscalls {
		if existing.Action != specs.ActAllow || len(existing.Args) != 0 {
			continue
		}
		if len(existing.Names) == 1 && existing.Names[0] == syscall {
			return true
		}
	}
	return false
}

// Returns an unconditional allow entry for the named syscall.
func unconditionalAllow(name string) specs.LinuxSyscall {
	return specs.LinuxSyscall{Names: []string{name}, Action: specs.ActAllow}
}

// Whether two syscall entries are structurally identical.
func syscallEqual(a, b specs.LinuxSyscall) bool {
	if a.Action != b.Action || len(a.Names) != len(b.Names) || len(a.Args) != len(b.Args) {
		return false
	}
	for i := range a.Names {
		if a.Names[i] != b.Names[i] {
			return false
		}
	}
	for i := range a.Args {
		if a.Args[i] != b.Args[i] {
			return false
		}
	}
	return true
}

// Returns a deep copy of a syscall entry.
func cloneSyscall(s specs.LinuxSyscall) specs.LinuxSyscall {
	out := specs.LinuxSyscall{
		Action:   s.Action,
		ErrnoRet: s.ErrnoRet,
		Names:    append([]string(nil), s.Names...),
	}
	if len(s.Args) != 0 {
		out.Args = append([]specs.LinuxSeccompArg(nil), s.Args...)
	}
	return out
}
