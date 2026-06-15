package seccomp

import (
	"errors"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
)

func nameArg(v string) agl.Arg { return agl.Arg{Type: agl.ArgName, Value: v} }

// Returns a Subsystem with an empty seccomp section.
func newSub() (*Subsystem, *specs.LinuxSeccomp) {
	s := &specs.LinuxSeccomp{}
	return New(s), s
}

func TestBuildIoctlSubFilterExpands(t *testing.T) {
	sub, s := newSub()
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("ioctl"), nameArg("tty")}}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(s.Syscalls) == 0 {
		t.Fatal("expected ioctl tty to expand to multiple entries")
	}
	for _, e := range s.Syscalls {
		if len(e.Names) != 1 || e.Names[0] != "ioctl" {
			t.Fatalf("unexpected entry: %+v", e)
		}
		if len(e.Args) != 1 || e.Args[0].Index != 1 || e.Args[0].Op != specs.OpEqualTo {
			t.Fatalf("bad arg shape: %+v", e.Args)
		}
	}
}

func TestBuildRejectsWhere(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "seccomp",
		Args: []agl.Arg{nameArg("read")}, Where: &agl.CompareExpr{}})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsKwargs(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "seccomp",
		Args: []agl.Arg{nameArg("read")}, Kwargs: []agl.Kwarg{{Key: "x", Value: nameArg("y")}}})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsTooManyArgs(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "seccomp",
		Args: []agl.Arg{nameArg("read"), nameArg("a"), nameArg("b")}})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsSubFilterOnArbitrarySyscall(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "seccomp",
		Args: []agl.Arg{nameArg("read"), nameArg("tty")}})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsUnknownSubFilter(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "seccomp",
		Args: []agl.Arg{nameArg("ioctl"), nameArg("bogus")}})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestNameReturnsSeccomp(t *testing.T) {
	sub, _ := newSub()
	if sub.Name() != subsystem.NameSeccomp {
		t.Fatalf("Name() = %q, want %q", sub.Name(), subsystem.NameSeccomp)
	}
}

func TestBuildRejectsUnknownSyscall(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("definitely_not_a_syscall")}})
	if !errors.Is(err, ErrUnknownSyscall) {
		t.Fatalf("err = %v, want ErrUnknownSyscall", err)
	}
}

func TestKeyReturnsFirstArg(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("read")}}
	if got := sub.Key(&g); got != "read" {
		t.Fatalf("Key() = %q, want %q", got, "read")
	}
}

func TestKeyEmptyWhenNoArgs(t *testing.T) {
	sub, _ := newSub()
	if got := sub.Key(&agl.Model{}); got != "" {
		t.Fatalf("Key() = %q, want empty", got)
	}
}

func TestBuildRejectsEmptyArgs(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "seccomp"})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonNameArg(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{{Type: agl.ArgInt, Value: "42"}}})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildAppendsMultiNameEntry(t *testing.T) {
	_, s := newSub()
	entry := specs.LinuxSyscall{
		Names:  []string{"read", "write"},
		Action: specs.ActAllow,
	}
	applyEntry(s, entry)
	if len(s.Syscalls) != 1 {
		t.Fatalf("expected 1 syscall entry, got %d", len(s.Syscalls))
	}
}

func TestReplaceWithUnconditionalDropsPrior(t *testing.T) {
	sub, s := newSub()
	// Add a filtered ioctl entry.
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("ioctl"), nameArg("tty")}}); err != nil {
		t.Fatalf("Build ioctl tty: %v", err)
	}
	priorCount := len(s.Syscalls)
	if priorCount == 0 {
		t.Fatal("expected at least one entry after ioctl tty")
	}
	// Now add unconditional ioctl — should replace all filtered entries.
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("ioctl")}}); err != nil {
		t.Fatalf("Build ioctl: %v", err)
	}
	// Count ioctl entries — should be exactly one (the unconditional one).
	ioctlCount := 0
	for _, sc := range s.Syscalls {
		if len(sc.Names) == 1 && sc.Names[0] == "ioctl" {
			ioctlCount++
		}
	}
	if ioctlCount != 1 {
		t.Fatalf("ioctl entries = %d, want 1 (unconditional replace)", ioctlCount)
	}
	unconditional := false
	for _, sc := range s.Syscalls {
		if len(sc.Names) == 1 && sc.Names[0] == "ioctl" && len(sc.Args) == 0 {
			unconditional = true
		}
	}
	if !unconditional {
		t.Fatal("expected unconditional ioctl entry")
	}
}

func TestEntryRedundantAfterUnconditional(t *testing.T) {
	sub, s := newSub()
	// First add unconditional read.
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("read")}}); err != nil {
		t.Fatalf("Build read: %v", err)
	}
	countBefore := len(s.Syscalls)
	// Build same syscall again — should be detected as redundant.
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("read")}}); err != nil {
		t.Fatalf("Build read again: %v", err)
	}
	if len(s.Syscalls) != countBefore {
		t.Fatalf("syscalls grew from %d to %d, expected no change (duplicate)", countBefore, len(s.Syscalls))
	}
}

func TestUnconditionalAllow(t *testing.T) {
	e := unconditionalAllow("read")
	if len(e.Names) != 1 || e.Names[0] != "read" {
		t.Fatalf("Names = %v, want [read]", e.Names)
	}
	if e.Action != specs.ActAllow {
		t.Fatalf("Action = %v, want ActAllow", e.Action)
	}
	if len(e.Args) != 0 {
		t.Fatalf("Args should be empty, got %v", e.Args)
	}
}

func TestSyscallEqualDiffArgCount(t *testing.T) {
	a := specs.LinuxSyscall{Names: []string{"read"}, Action: specs.ActAllow}
	b := specs.LinuxSyscall{Names: []string{"read"}, Action: specs.ActAllow,
		Args: []specs.LinuxSeccompArg{{Index: 0, Value: 1, Op: specs.OpEqualTo}},
	}
	if syscallEqual(a, b) {
		t.Fatal("expected not equal when arg counts differ")
	}
}

func TestBuildFcntlSubFilter(t *testing.T) {
	sub, s := newSub()
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("fcntl"), nameArg("flags")}}); err != nil {
		t.Fatalf("Build fcntl flags: %v", err)
	}
	if len(s.Syscalls) == 0 {
		t.Fatal("expected fcntl flags to expand to entries")
	}
	for _, e := range s.Syscalls {
		if len(e.Names) != 1 || e.Names[0] != "fcntl" {
			t.Fatalf("unexpected entry: %+v", e)
		}
	}
}

func TestBuildPrctlSubFilter(t *testing.T) {
	sub, s := newSub()
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("prctl"), nameArg("name")}}); err != nil {
		t.Fatalf("Build prctl name: %v", err)
	}
	if len(s.Syscalls) == 0 {
		t.Fatal("expected prctl name to expand to entries")
	}
	for _, e := range s.Syscalls {
		if len(e.Names) != 1 || e.Names[0] != "prctl" {
			t.Fatalf("unexpected entry: %+v", e)
		}
	}
}

func TestBuildConditionalEntryDuplicate(t *testing.T) {
	sub, s := newSub()
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("ioctl"), nameArg("tty")}}); err != nil {
		t.Fatalf("first Build ioctl tty: %v", err)
	}
	countAfterFirst := len(s.Syscalls)
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("ioctl"), nameArg("tty")}}); err != nil {
		t.Fatalf("second Build ioctl tty: %v", err)
	}
	if len(s.Syscalls) != countAfterFirst {
		t.Fatalf("syscalls grew from %d to %d, expected no change (duplicate filtered entry)", countAfterFirst, len(s.Syscalls))
	}
}

func TestReplaceWithUnconditionalKeepsOtherSyscalls(t *testing.T) {
	sub, s := newSub()
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("read")}}); err != nil {
		t.Fatalf("Build read: %v", err)
	}
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("ioctl"), nameArg("tty")}}); err != nil {
		t.Fatalf("Build ioctl tty: %v", err)
	}
	// Replace filtered ioctl entries with unconditional. The read entry must survive.
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("ioctl")}}); err != nil {
		t.Fatalf("Build ioctl unconditional: %v", err)
	}
	hasRead := false
	for _, e := range s.Syscalls {
		if len(e.Names) == 1 && e.Names[0] == "read" {
			hasRead = true
			break
		}
	}
	if !hasRead {
		t.Fatal("read entry should be preserved after unconditional ioctl replace")
	}
}

func TestEntryRedundantUnconditionalCoversConditional(t *testing.T) {
	sub, s := newSub()
	// Add unconditional ioctl first.
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("ioctl")}}); err != nil {
		t.Fatalf("Build ioctl: %v", err)
	}
	countBefore := len(s.Syscalls)
	// Now try to add a conditional ioctl+tty entry — should be redundant.
	if err := sub.Build(&agl.Model{Subsystem: "seccomp", Args: []agl.Arg{nameArg("ioctl"), nameArg("tty")}}); err != nil {
		t.Fatalf("Build ioctl tty: %v", err)
	}
	if len(s.Syscalls) != countBefore {
		t.Fatalf("syscalls grew from %d to %d, expected no change (unconditional covers conditional)", countBefore, len(s.Syscalls))
	}
}

func TestSyscallEqualArgsEqual(t *testing.T) {
	arg := specs.LinuxSeccompArg{Index: 0, Value: 42, Op: specs.OpEqualTo}
	a := specs.LinuxSyscall{Names: []string{"read"}, Action: specs.ActAllow, Args: []specs.LinuxSeccompArg{arg}}
	b := specs.LinuxSyscall{Names: []string{"read"}, Action: specs.ActAllow, Args: []specs.LinuxSeccompArg{arg}}
	if !syscallEqual(a, b) {
		t.Fatal("expected equal entries to be equal")
	}
}

func TestSyscallEqualArgsDiffer(t *testing.T) {
	a := specs.LinuxSyscall{Names: []string{"read"}, Action: specs.ActAllow,
		Args: []specs.LinuxSeccompArg{{Index: 0, Value: 1, Op: specs.OpEqualTo}},
	}
	b := specs.LinuxSyscall{Names: []string{"read"}, Action: specs.ActAllow,
		Args: []specs.LinuxSeccompArg{{Index: 0, Value: 2, Op: specs.OpEqualTo}},
	}
	if syscallEqual(a, b) {
		t.Fatal("expected entries with different arg values to be not equal")
	}
}
