package cgroup

import (
	"errors"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/internal/manifest/grant"
	"github.com/cruciblehq/crux/internal/runtime/shared"
)

// Returns a Subsystem bound to a fresh LinuxResources for assertion.
func newSub() (*Subsystem, *specs.LinuxResources) {
	lr := &specs.LinuxResources{}
	return New(lr), lr
}

// Parses src as a grant and feeds it into sub.Build.
func buildSrc(t *testing.T, sub *Subsystem, src string) error {
	t.Helper()
	g, err := grant.Parse(src)
	if err != nil {
		t.Fatalf("Parse(%q): %v", src, err)
	}
	return sub.Build(*g)
}

// Asserts that lr.Unified[knob] equals want.
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

func TestSubsystemName(t *testing.T) {
	sub, _ := newSub()
	if got := sub.Name(); got != shared.NameCgroup {
		t.Fatalf("Name() = %v, want %v", got, shared.NameCgroup)
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
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestBuildIdempotent(t *testing.T) {
	sub, lr := newSub()
	for range 2 {
		if err := buildSrc(t, sub, ".cgroup cpu.weight 100"); err != nil {
			t.Fatal(err)
		}
	}
	wantUnified(t, lr, "cpu.weight", "100")
}

func TestBuildUnknownKnob(t *testing.T) {
	sub, _ := newSub()
	err := buildSrc(t, sub, ".cgroup not.a.knob 1")
	if !errors.Is(err, ErrUnknownKnob) {
		t.Fatalf("err = %v, want ErrUnknownKnob", err)
	}
}

func TestBuildRejectsWhere(t *testing.T) {
	sub, _ := newSub()
	if err := buildSrc(t, sub, ".cgroup cpu.weight 100 where 1 = 1"); err == nil {
		t.Fatal("expected error for where clause")
	}
}

func TestBuildMissingValue(t *testing.T) {
	sub, _ := newSub()
	if err := buildSrc(t, sub, ".cgroup cpu.weight"); err == nil {
		t.Fatal("expected error for missing value")
	}
}

func TestBuildMissingKnob(t *testing.T) {
	sub, _ := newSub()
	g := grant.Grant{Subsystem: "cgroup"}
	err := sub.Build(g)
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
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
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildKwargsJoinedAfterPositional(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup io.max 8 0 rbps=1048576"); err != nil {
		t.Fatal(err)
	}
	wantUnified(t, lr, "io.max", "8 0 rbps=1048576")
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

func TestMergeReplaysUnifiedThroughBuild(t *testing.T) {
	sub, lr := newSub()
	src := shared.Spec{OCI: &specs.Spec{Linux: &specs.Linux{Resources: &specs.LinuxResources{
		Unified: map[string]string{
			"cpu.weight": "300",
			"memory.max": "1048576",
		},
	}}}}
	if err := sub.Merge(src); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	wantUnified(t, lr, "cpu.weight", "300")
	wantUnified(t, lr, "memory.max", "1048576")
}

func TestMergeReplaysMultilineListKnob(t *testing.T) {
	sub, lr := newSub()
	src := shared.Spec{OCI: &specs.Spec{Linux: &specs.Linux{Resources: &specs.LinuxResources{
		Unified: map[string]string{
			"io.max": "8 0 rbps=1000\n8 16 rbps=2000",
		},
	}}}}
	if err := sub.Merge(src); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	got := lr.Unified["io.max"]
	want := "8 0 rbps=1000\n8 16 rbps=2000"
	if got != want {
		t.Fatalf("io.max = %q, want %q", got, want)
	}
}

func TestMergeNilSourceIsNoOp(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup cpu.weight 100"); err != nil {
		t.Fatal(err)
	}
	if err := sub.Merge(shared.Spec{}); err != nil {
		t.Fatalf("Merge(empty): %v", err)
	}
	wantUnified(t, lr, "cpu.weight", "100")
}

func TestMergeConflict(t *testing.T) {
	sub, _ := newSub()
	if err := buildSrc(t, sub, ".cgroup cpu.weight 100"); err != nil {
		t.Fatal(err)
	}
	src := shared.Spec{OCI: &specs.Spec{Linux: &specs.Linux{Resources: &specs.LinuxResources{
		Unified: map[string]string{"cpu.weight": "200"},
	}}}}
	err := sub.Merge(src)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestBuildValueJoinsArgsAndKwargs(t *testing.T) {
	args := []grant.Arg{{Type: grant.ArgInt, Value: "8"}, {Type: grant.ArgInt, Value: "0"}}
	kwargs := []grant.Kwarg{{Key: "rbps", Value: grant.Arg{Type: grant.ArgInt, Value: "1024"}}}
	if got, want := buildValue(args, kwargs), "8 0 rbps=1024"; got != want {
		t.Fatalf("buildValue = %q, want %q", got, want)
	}
}

func TestBuildValueEmpty(t *testing.T) {
	if got := buildValue(nil, nil); got != "" {
		t.Fatalf("buildValue(nil,nil) = %q, want empty", got)
	}
}

func TestBuildDeviceFlushesTypedList(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup devices c 8 0 rw"); err != nil {
		t.Fatal(err)
	}
	if _, ok := lr.Unified["devices"]; ok {
		t.Fatalf("devices should not appear in Unified: %v", lr.Unified)
	}
	if len(lr.Devices) != 1 {
		t.Fatalf("Devices = %+v, want 1 entry", lr.Devices)
	}
	d := lr.Devices[0]
	if !d.Allow || d.Type != "c" || d.Major == nil || *d.Major != 8 || d.Minor == nil || *d.Minor != 0 || d.Access != "rw" {
		t.Fatalf("device = %+v", d)
	}
}

func TestMergeReplaysDeviceEntries(t *testing.T) {
	sub, lr := newSub()
	maj, min := int64(8), int64(16)
	src := shared.Spec{OCI: &specs.Spec{Linux: &specs.Linux{Resources: &specs.LinuxResources{
		Devices: []specs.LinuxDeviceCgroup{
			{Allow: true, Type: "b", Major: &maj, Minor: &min, Access: "r"},
			{Allow: false, Type: "a"}, // deny entries are skipped
		},
	}}}}
	if err := sub.Merge(src); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if len(lr.Devices) != 1 {
		t.Fatalf("Devices = %+v, want 1 entry", lr.Devices)
	}
	got := lr.Devices[0]
	if got.Type != "b" || *got.Major != 8 || *got.Minor != 16 || got.Access != "r" {
		t.Fatalf("device = %+v", got)
	}
}

func TestMergeReplaysDeviceWithDefaultType(t *testing.T) {
	sub, lr := newSub()
	src := shared.Spec{OCI: &specs.Spec{Linux: &specs.Linux{Resources: &specs.LinuxResources{
		Devices: []specs.LinuxDeviceCgroup{{Allow: true, Access: "rwm"}},
	}}}}
	if err := sub.Merge(src); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if len(lr.Devices) != 1 || lr.Devices[0].Type != "a" {
		t.Fatalf("Devices = %+v", lr.Devices)
	}
}

func TestBuildCPUSetCPUsIndexList(t *testing.T) {
	sub, lr := newSub()
	if err := buildSrc(t, sub, ".cgroup cpuset.cpus \"0-2,4\""); err != nil {
		t.Fatal(err)
	}
	wantUnified(t, lr, "cpuset.cpus", "0-2,4")
}

func TestBuildListFieldDispatch(t *testing.T) {
	cases := []struct {
		name string
		src  string
		knob string
		want string
	}{
		{"hugetlb", ".cgroup hugetlb \"2MB\" max=1048576", "hugetlb", "2MB max=1048576"},
		{"rdma", ".cgroup rdma mlx5_0 hca_handle=4 hca_object=16", "rdma", "mlx5_0 hca_handle=4 hca_object=16"},
		{"misc", ".cgroup misc sev max=10", "misc", "sev max=10"},
		{"io.max", ".cgroup io.max 8 0 rbps=1024", "io.max", "8 0 rbps=1024"},
		{"io.latency", ".cgroup io.latency 8 0 target=2000", "io.latency", "8 0 target=2000"},
		{"io.cost.model", ".cgroup io.cost.model 8 0 rbps=1000", "io.cost.model", "8 0 rbps=1000"},
		{"io.cost.qos", ".cgroup io.cost.qos 8 0 enable=true", "io.cost.qos", "8 0 enable=true"},
		{"io.weight per device", ".cgroup io.weight \"8:16\" 200", "io.weight", "8:16 200"},
		{"cgroup.subtree_control", ".cgroup cgroup.subtree_control cpu memory", "cgroup.subtree_control", "cpu memory"},
		{"dmem.max", ".cgroup dmem.max gpu0 1024", "dmem.max", "gpu0 1024"},
		{"psi.io", ".cgroup psi.io some 80 1000000", "psi.io", "some 80 1000000"},
		{"psi.irq", ".cgroup psi.irq full 60 2000000", "psi.irq", "full 60 2000000"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sub, lr := newSub()
			if err := buildSrc(t, sub, c.src); err != nil {
				t.Fatal(err)
			}
			wantUnified(t, lr, c.knob, c.want)
		})
	}
}
