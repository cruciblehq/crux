package seccomp

import (
	"errors"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/internal/manifest/grant"
	"github.com/cruciblehq/crux/internal/runtime/shared"
)

func nameArg(v string) grant.Arg { return grant.Arg{Type: grant.ArgName, Value: v} }

// Returns a Subsystem with an empty seccomp section.
func newSub() (*Subsystem, *specs.LinuxSeccomp) {
	s := &specs.LinuxSeccomp{}
	return New(s), s
}

// Wraps a seccomp section into a unified spec for Merge inputs.
func wrap(s *specs.LinuxSeccomp) shared.Spec {
	return shared.Spec{OCI: &specs.Spec{Linux: &specs.Linux{Seccomp: s}}}
}

func TestBuildIoctlSubFilterExpands(t *testing.T) {
	sub, s := newSub()
	if err := sub.Build(grant.Grant{Subsystem: "seccomp", Args: []grant.Arg{nameArg("ioctl"), nameArg("tty")}}); err != nil {
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

func TestBuildUnconditionalSubsumesFiltered(t *testing.T) {
	sub, s := newSub()
	if err := sub.Build(grant.Grant{Subsystem: "seccomp", Args: []grant.Arg{nameArg("ioctl"), nameArg("tty")}}); err != nil {
		t.Fatal(err)
	}
	if err := sub.Build(grant.Grant{Subsystem: "seccomp", Args: []grant.Arg{nameArg("ioctl")}}); err != nil {
		t.Fatal(err)
	}
	if len(s.Syscalls) != 1 {
		t.Fatalf("len = %d, want 1 after subsumption", len(s.Syscalls))
	}
	if len(s.Syscalls[0].Args) != 0 {
		t.Fatal("unconditional should have no args")
	}
}

func TestBuildFilteredAfterUnconditionalIsRedundant(t *testing.T) {
	sub, s := newSub()
	if err := sub.Build(grant.Grant{Subsystem: "seccomp", Args: []grant.Arg{nameArg("ioctl")}}); err != nil {
		t.Fatal(err)
	}
	if err := sub.Build(grant.Grant{Subsystem: "seccomp", Args: []grant.Arg{nameArg("ioctl"), nameArg("tty")}}); err != nil {
		t.Fatal(err)
	}
	if len(s.Syscalls) != 1 {
		t.Fatalf("len = %d, want 1", len(s.Syscalls))
	}
}

func TestBuildIdempotent(t *testing.T) {
	sub, s := newSub()
	g := grant.Grant{Subsystem: "seccomp", Args: []grant.Arg{nameArg("read")}}
	if err := sub.Build(g); err != nil {
		t.Fatal(err)
	}
	if err := sub.Build(g); err != nil {
		t.Fatal(err)
	}
	if len(s.Syscalls) != 1 {
		t.Fatalf("len = %d, want 1", len(s.Syscalls))
	}
}

func TestBuildRejectsWhere(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(grant.Grant{Subsystem: "seccomp",
		Args: []grant.Arg{nameArg("read")}, Where: &grant.CompareExpr{}})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsKwargs(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(grant.Grant{Subsystem: "seccomp",
		Args: []grant.Arg{nameArg("read")}, Kwargs: []grant.Kwarg{{Key: "x", Value: nameArg("y")}}})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsTooManyArgs(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(grant.Grant{Subsystem: "seccomp",
		Args: []grant.Arg{nameArg("read"), nameArg("a"), nameArg("b")}})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsSubFilterOnArbitrarySyscall(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(grant.Grant{Subsystem: "seccomp",
		Args: []grant.Arg{nameArg("read"), nameArg("tty")}})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsUnknownSubFilter(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(grant.Grant{Subsystem: "seccomp",
		Args: []grant.Arg{nameArg("ioctl"), nameArg("bogus")}})
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestMergeUnionsAndSubsumes(t *testing.T) {
	dstSub, dst := newSub()
	if err := dstSub.Build(grant.Grant{Subsystem: "seccomp", Args: []grant.Arg{nameArg("ioctl"), nameArg("tty")}}); err != nil {
		t.Fatal(err)
	}
	srcSub, src := newSub()
	if err := srcSub.Build(grant.Grant{Subsystem: "seccomp", Args: []grant.Arg{nameArg("ioctl")}}); err != nil {
		t.Fatal(err)
	}
	if err := dstSub.Merge(wrap(src)); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if len(dst.Syscalls) != 1 || len(dst.Syscalls[0].Args) != 0 {
		t.Fatalf("expected single unconditional after merge, got %+v", dst.Syscalls)
	}
}

func TestMergeNilIsNoOp(t *testing.T) {
	sub, s := newSub()
	if err := sub.Build(grant.Grant{Subsystem: "seccomp", Args: []grant.Arg{nameArg("read")}}); err != nil {
		t.Fatal(err)
	}
	if err := sub.Merge(shared.Spec{}); err != nil {
		t.Fatalf("Merge(empty): %v", err)
	}
	if len(s.Syscalls) != 1 {
		t.Fatal("Merge(empty) mutated state")
	}
}

func TestNameReturnsSeccomp(t *testing.T) {
	sub, _ := newSub()
	if sub.Name() != shared.NameSeccomp {
		t.Fatalf("Name() = %q, want %q", sub.Name(), shared.NameSeccomp)
	}
}

func TestBuildRejectsUnknownSyscall(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(grant.Grant{Subsystem: "seccomp", Args: []grant.Arg{nameArg("definitely_not_a_syscall")}})
	if !errors.Is(err, ErrUnknownSyscall) {
		t.Fatalf("err = %v, want ErrUnknownSyscall", err)
	}
}
