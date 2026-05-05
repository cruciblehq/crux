package runtime

import (
	"errors"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/internal/aegis"
)

func mustParse(t *testing.T, src string) *aegis.Model {
	t.Helper()
	p, err := aegis.Parse(src)
	if err != nil {
		t.Fatalf("Parse(%q): %v", src, err)
	}
	return p
}

// Returns true if the seccomp spec allows the named syscall via an
// unconditional allow entry.
func seccompAllows(seccomp *specs.LinuxSeccomp, name string) bool {
	if seccomp == nil {
		return false
	}
	for _, sc := range seccomp.Syscalls {
		if sc.Action != specs.ActAllow {
			continue
		}
		for _, n := range sc.Names {
			if n == name {
				return true
			}
		}
	}
	return false
}

// Returns the soft limit for the named rlimit type, or 0 if absent.
func rlimitSoft(rlimits []specs.POSIXRlimit, typ string) uint64 {
	for _, r := range rlimits {
		if r.Type == typ {
			return r.Soft
		}
	}
	return 0
}

func TestBuilderRoutesByName(t *testing.T) {
	b := NewBuilder()
	if err := b.Build(mustParse(t, ".cap net_admin")); err != nil {
		t.Fatal(err)
	}
	if err := b.Build(mustParse(t, ".rlimit nofile 1024")); err != nil {
		t.Fatal(err)
	}
	c := b.Spec()
	if c == nil {
		t.Fatal("spec = nil")
	}
	caps := c.OCI.Process.Capabilities
	if caps == nil || len(caps.Bounding) == 0 {
		t.Fatal("cap state missing")
	}
	if len(c.OCI.Process.Rlimits) == 0 {
		t.Fatal("rlimit state missing")
	}
	if soft := rlimitSoft(c.OCI.Process.Rlimits, "RLIMIT_NOFILE"); soft != 1024 {
		t.Fatalf("RLIMIT_NOFILE soft = %d, want 1024", soft)
	}
}

func TestBuilderUnknownSubsystem(t *testing.T) {
	b := NewBuilder()
	g := mustParse(t, ".cap net_admin")
	g.Subsystem = "bogus"
	err := b.Build(g)
	if err == nil || !errors.Is(err, ErrUnknownSubsystem) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuilderConfigEmptyWhenNoGrants(t *testing.T) {
	b := NewBuilder()
	c := b.Spec()
	if c == nil {
		t.Fatal("spec = nil, want non-nil all-deny snapshot")
	}
	if c.OCI == nil {
		t.Fatal("oci = nil, want non-nil baseline")
	}
	if caps := c.OCI.Process.Capabilities; caps == nil || len(caps.Bounding) != 0 {
		t.Fatalf("cap = %+v, want empty", caps)
	}
	checkSeccompBaseline(t, c.OCI.Linux.Seccomp)
	if c.MAC == nil || len(c.MAC.Rules) != 0 {
		t.Fatalf("mac = %+v, want empty", c.MAC)
	}
	if c.Fcap == nil || len(c.Fcap.Entries) != 0 {
		t.Fatalf("fcap = %+v, want empty", c.Fcap)
	}
	if c.OCI.Linux.Resources == nil {
		t.Fatal("oci.Linux.Resources = nil, want non-nil baseline")
	}
	if u := c.OCI.Linux.Resources.Unified; len(u) != 0 {
		t.Fatalf("oci.Linux.Resources.Unified = %v, want empty", u)
	}
}

// Asserts that the seccomp section matches the deny-all baseline:
// non-nil and the only allow entry permits exit_group.
func checkSeccompBaseline(t *testing.T, seccomp *specs.LinuxSeccomp) {
	t.Helper()
	if seccomp == nil {
		t.Fatal("seccomp = nil")
	}
	for _, sc := range seccomp.Syscalls {
		if sc.Action != specs.ActAllow {
			continue
		}
		if len(sc.Names) != 1 || sc.Names[0] != "exit_group" {
			t.Fatalf("unexpected baseline allow entry: %+v", sc)
		}
	}
}

func TestBuilderMergeNil(t *testing.T) {
	b := NewBuilder()
	if err := b.Merge(nil); err != nil {
		t.Fatal(err)
	}
}

func TestBuilderMergeFoldsAllSubsystems(t *testing.T) {
	src := NewBuilder()
	for _, g := range []string{
		".cap net_admin",
		".rlimit nofile 2048",
		".seccomp read",
		".fcap net_raw effective \"/usr/bin/ping\"",
		".mac file_open where task.uid = 0",
		".cgroup memory.max 1073741824",
	} {
		if err := src.Build(mustParse(t, g)); err != nil {
			t.Fatalf("src.Build(%q): %v", g, err)
		}
	}

	dst := NewBuilder()
	if err := dst.Merge(src.Spec()); err != nil {
		t.Fatalf("Merge: %v", err)
	}

	c := dst.Spec()
	if caps := c.OCI.Process.Capabilities; caps == nil || len(caps.Bounding) == 0 {
		t.Fatal("cap section not merged")
	}
	if rlimitSoft(c.OCI.Process.Rlimits, "RLIMIT_NOFILE") != 2048 {
		t.Fatalf("rlimit not merged: %+v", c.OCI.Process.Rlimits)
	}
	if !seccompAllows(c.OCI.Linux.Seccomp, "read") {
		t.Fatal("seccomp section not merged")
	}
	if c.Fcap == nil || len(c.Fcap.Entries) == 0 {
		t.Fatal("fcap section not merged")
	}
	if c.MAC == nil || len(c.MAC.Rules) == 0 {
		t.Fatal("mac section not merged")
	}
	if c.OCI.Linux.Resources == nil || len(c.OCI.Linux.Resources.Unified) == 0 {
		t.Fatal("cgroup section not merged")
	}
}
