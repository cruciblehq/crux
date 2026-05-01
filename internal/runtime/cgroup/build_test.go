package cgroup

import (
	"errors"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/internal/manifest/grant"
)

func newSub() (*Subsystem, *specs.LinuxResources) {
	lr := &specs.LinuxResources{}
	return New(lr), lr
}

func buildSrc(t *testing.T, sub *Subsystem, src string) error {
	t.Helper()
	g, err := grant.Parse(src)
	if err != nil {
		t.Fatalf("Parse(%q): %v", src, err)
	}
	return sub.Build(*g)
}

func wantUnified(t *testing.T, lr *specs.LinuxResources, knob, want string) {
	t.Helper()
	got, ok := lr.Unified[knob]
	if !ok {
		t.Fatalf("unified[%q] missing; have %v", knob, lr.Unified)
	}
	if got != want {
		t.Fatalf("unified[%q] = %q, want %q", knob, got, want)
	}
}

func TestBuildScalarKnob(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup cpu.weight 100"); err != nil {
		t.Fatal(err)
	}
	wantUnified(t, lr, "cpu.weight", "100")
}

func TestBuildScalarConflict(t *testing.T) {
	sub, _ := newSub()
	if err := buildSrc(t, sub, ".cgroup cpu.weight 100"); err != nil {
		t.Fatal(err)
	}
	err := buildSrc(t, sub, ".cgroup cpu.weight 200")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want errConflict", err)
	}
}

func TestBuildIdempotent(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup cpu.weight 100"); err != nil {
		t.Fatal(err)
	}
	if err := buildSrc(t, sub, ".cgroup cpu.weight 100"); err != nil {
		t.Fatal(err)
	}
	wantUnified(t, lr, "cpu.weight", "100")
}

func TestBuildUnknownKnob(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, ".cgroup not.a.knob 1")
	if !errors.Is(err, ErrUnknownKnob) {
		t.Fatalf("err = %v, want errUnknownKnob", err)
	}
}

func TestBuildCompositeKnobIOMax(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup io.max 8 0 rbps=1048576"); err != nil {
		t.Fatal(err)
	}
	wantUnified(t, lr, "io.max", "8 0 rbps=1048576")
}

func TestBuildQuotedStringValue(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup cpuset.cpus \"0-3\""); err != nil {
		t.Fatal(err)
	}
	wantUnified(t, lr, "cpuset.cpus", "0-3")
}

func TestBuildRejectsWhere(t *testing.T) {
	sub, _ := newSub()
	if err := buildSrc(t, sub, ".cgroup cpu.weight 100 where 1 = 1"); err == nil {
		t.Fatal("expected error for where")
	}
}

func TestBuildMissingValue(t *testing.T) {
	sub, _ := newSub()
	if err := buildSrc(t, sub, ".cgroup cpu.weight"); err == nil {
		t.Fatal("expected error for missing value")
	}
}

func TestBuildNonNameKnobArg(t *testing.T) {
	g, err := grant.Parse(".cgroup cpu.weight 100")
	if err != nil {
		t.Fatal(err)
	}
	g.Args[0] = grant.Arg{Type: grant.ArgInt, Value: "42"}
	sub, _ := newSub()
	if err := sub.Build(*g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want errInvalidGrant", err)
	}
}

func TestSpecEmptyWhenNoGrants(t *testing.T) {
	if got := newSpec(); got == nil {
		t.Fatal("newSpec() = nil")
	}
}

func TestBuildMemoryMax(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup memory.max 104857600"); err != nil {
		t.Fatal(err)
	}
	wantUnified(t, lr, "memory.max", "104857600")
}

func TestBuildMemoryHigh(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup memory.high 52428800"); err != nil {
		t.Fatal(err)
	}
	wantUnified(t, lr, "memory.high", "52428800")
}

func TestBuildPidsMax(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup pids.max 64"); err != nil {
		t.Fatal(err)
	}
	wantUnified(t, lr, "pids.max", "64")
}

func TestBuildCPUBurst(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup cpu.burst 5000"); err != nil {
		t.Fatal(err)
	}
	wantUnified(t, lr, "cpu.burst", "5000")
}

func TestBuildSubtreeControl(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup cgroup.subtree_control cpu memory"); err != nil {
		t.Fatal(err)
	}
	wantUnified(t, lr, "cgroup.subtree_control", "cpu memory")
}

func TestBuildMultipleKnobs(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup cpu.weight 200"); err != nil {
		t.Fatal(err)
	}
	if err := buildSrc(t, sub, ".cgroup memory.max 67108864"); err != nil {
		t.Fatal(err)
	}
	wantUnified(t, lr, "cpu.weight", "200")
	wantUnified(t, lr, "memory.max", "67108864")
}
