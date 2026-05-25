package seccomp

import (
	"errors"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/resource/affordance/agl"
	"github.com/cruciblehq/crux/resource/affordance/subsystem"
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
