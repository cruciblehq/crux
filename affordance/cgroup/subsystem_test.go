package cgroup

import (
	"errors"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
)

// Returns a Subsystem bound to a new LinuxResources for assertion.
func newSub() (*Subsystem, *specs.LinuxResources) {
	lr := &specs.LinuxResources{}
	return New(lr), lr
}

// Parses src as a grant and feeds it into sub.Build.
func buildSrc(t *testing.T, sub *Subsystem, src string) error {
	t.Helper()
	g, err := agl.Parse(src)
	if err != nil {
		t.Fatalf("Parse(%q): %v", src, err)
	}
	return sub.Build(g)
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
	if got := sub.Name(); got != subsystem.NameCgroup {
		t.Fatalf("Name() = %v, want %v", got, subsystem.NameCgroup)
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

func TestBuildSameKnobConflict(t *testing.T) {
	sub, _ := newSub()
	if err := buildSrc(t, sub, ".cgroup cpu.weight 100"); err != nil {
		t.Fatal(err)
	}
	err := buildSrc(t, sub, ".cgroup cpu.weight 100")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
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
	g := agl.Model{Subsystem: "cgroup"}
	err := sub.Build(&g)
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildNonNameKnobArg(t *testing.T) {
	g := agl.Model{
		Subsystem: "cgroup",
		Args: []agl.Arg{
			{Type: agl.ArgInt, Value: "42"},
			{Type: agl.ArgInt, Value: "100"},
		},
	}
	sub, _ := newSub()
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
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

func TestBuildValueJoinsArgsAndKwargs(t *testing.T) {
	args := []agl.Arg{{Type: agl.ArgInt, Value: "8"}, {Type: agl.ArgInt, Value: "0"}}
	kwargs := []agl.Kwarg{{Key: "rbps", Value: agl.Arg{Type: agl.ArgInt, Value: "1024"}}}
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
		{"hugetlb", ".cgroup hugetlb \"2MB\" max=1048576", "hugetlb.2MB.max", "1048576"},
		{"rdma.max", ".cgroup rdma.max mlx5_0 hca_handle=4 hca_object=16", "rdma.max", "mlx5_0 hca_handle=4 hca_object=16"},
		{"misc.max", ".cgroup misc.max sev max=10", "misc.max", "sev max=10"},
		{"io.max", ".cgroup io.max 8 0 rbps=1024", "io.max", "8 0 rbps=1024"},
		{"io.latency", ".cgroup io.latency 8 0 target=2000", "io.latency", "8 0 target=2000"},
		{"io.cost.model", ".cgroup io.cost.model 8 0 rbps=1000", "io.cost.model", "8 0 rbps=1000"},
		{"io.cost.qos", ".cgroup io.cost.qos 8 0 enable=true", "io.cost.qos", "8 0 enable=true"},
		{"io.weight per device", ".cgroup io.weight \"8:16\" 200", "io.weight", "8:16 200"},
		{"cgroup.subtree_control", ".cgroup cgroup.subtree_control cpu memory", "cgroup.subtree_control", "cpu memory"},
		{"dmem.max", ".cgroup dmem.max gpu0 1024", "dmem.max", "gpu0 1024"},
		{"io.pressure", ".cgroup io.pressure some 80 1000000", "io.pressure", "some 80 1000000"},
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

func TestKey(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		// Scalar knobs: key = knob name.
		{"cpu.weight scalar", ".cgroup cpu.weight 100", "cpu.weight"},
		{"cpu.max scalar", ".cgroup cpu.max 50000 100000", "cpu.max"},
		{"memory.max scalar", ".cgroup memory.max 67108864", "memory.max"},
		{"pids.max scalar", ".cgroup pids.max 100", "pids.max"},
		{"cgroup.subtree_control", ".cgroup cgroup.subtree_control cpu memory", "cgroup.subtree_control"},

		// MAJOR MINOR list knobs: key = knob:major:minor.
		{"io.max", ".cgroup io.max 8 0 rbps=1024", "io.max:8:0"},
		{"io.latency", ".cgroup io.latency 8 0 target=2000", "io.latency:8:0"},
		{"io.cost.model", ".cgroup io.cost.model 8 0 rbps=1000", "io.cost.model:8:0"},
		{"io.cost.qos", ".cgroup io.cost.qos 8 0 enable=true", "io.cost.qos:8:0"},

		// io.weight: scalar key vs per-device key.
		{"io.weight scalar", ".cgroup io.weight 500", "io.weight"},
		{"io.weight per-device", ".cgroup io.weight \"8:16\" 200", "io.weight:8:16"},

		// Single-token identity list knobs: key = knob:token.
		{"hugetlb", ".cgroup hugetlb \"2MB\" max=1048576", "hugetlb:2MB"},
		{"rdma.max", ".cgroup rdma.max mlx5_0 hca_handle=4 hca_object=16", "rdma.max:mlx5_0"},
		{"misc.max", ".cgroup misc.max sev max=10", "misc.max:sev"},
		{"dmem.max", ".cgroup dmem.max gpu0 1024", "dmem.max:gpu0"},
		{"dmem.min", ".cgroup dmem.min gpu0 512", "dmem.min:gpu0"},
		{"cpu.pressure", ".cgroup cpu.pressure some 80 1000000", "cpu.pressure:some"},
		{"memory.pressure", ".cgroup memory.pressure full 60 2000000", "memory.pressure:full"},
		{"io.pressure", ".cgroup io.pressure some 80 1000000", "io.pressure:some"},

		// Device knobs: identity is type + major + minor; key = "devices:type:major:minor".
		{"devices", ".cgroup devices c 8 0 rw", "devices:c:8:0"},
	}
	sub, _ := newSub()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g, err := agl.Parse(c.src)
			if err != nil {
				t.Fatalf("Parse(%q): %v", c.src, err)
			}
			if got := sub.Key(g); got != c.want {
				t.Fatalf("Key() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestKeyEmptyModel(t *testing.T) {
	sub, _ := newSub()
	if got := sub.Key(&agl.Model{}); got != "" {
		t.Fatalf("Key() on empty model = %q, want empty", got)
	}
}

// Two grants for the same device identity (type + major + minor) must conflict
// even when they specify different access bits. Callers must combine all
// access permissions into a single grant.
func TestBuildDeviceSameIdentityConflicts(t *testing.T) {
	sub, _ := newSub()
	if err := buildSrc(t, sub, ".cgroup devices c 8 0 r"); err != nil {
		t.Fatal(err)
	}
	if err := buildSrc(t, sub, ".cgroup devices c 8 0 w"); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}
