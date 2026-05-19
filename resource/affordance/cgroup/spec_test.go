package cgroup

import (
	"reflect"
	"testing"
)

func TestNewSpecInitialisesSetMap(t *testing.T) {
	s := newSpec()
	if s.set == nil {
		t.Fatal("set map is nil")
	}
}

func TestSpecTopLevelFieldsAreTagged(t *testing.T) {
	t1 := reflect.TypeFor[spec]()
	for i := range t1.NumField() {
		f := t1.Field(i)
		if f.Name == "set" { // tracker; not a knob target.
			continue
		}
		if f.Tag.Get(knobStructTag) == "" {
			t.Errorf("field %q missing knob tag", f.Name)
		}
	}
}

func TestSpecTopLevelKnobTagsMatchExpectedPrefixes(t *testing.T) {
	want := map[string]string{
		"Cgroup":  "cgroup",
		"CPU":     "cpu",
		"Memory":  "memory",
		"IO":      "io",
		"PIDs":    "pids",
		"CPUSet":  "cpuset",
		"PSI":     "psi",
		"HugeTLB": "hugetlb",
		"RDMA":    "rdma",
		"Misc":    "misc",
		"Devices": "devices",
		"Dmem":    "dmem",
	}
	t1 := reflect.TypeFor[spec]()
	for name, tag := range want {
		f, ok := t1.FieldByName(name)
		if !ok {
			t.Errorf("field %q not found", name)
			continue
		}
		if got := f.Tag.Get(knobStructTag); got != tag {
			t.Errorf("field %q knob tag = %q, want %q", name, got, tag)
		}
	}
}

func TestSpecSetFieldExcludedFromJSON(t *testing.T) {
	f, ok := reflect.TypeFor[spec]().FieldByName("set")
	if !ok {
		t.Fatal("set field missing")
	}
	if got := f.Tag.Get("json"); got != "-" {
		t.Fatalf("set json tag = %q, want %q", got, "-")
	}
}

func TestNewSpecAppliesDefaults(t *testing.T) {
	s := newSpec()
	if s.CPU.Period != 100000 {
		t.Errorf("cpu.period = %d, want 100000", s.CPU.Period)
	}
	if s.CPU.Weight != 1 {
		t.Errorf("cpu.weight = %d, want 1", s.CPU.Weight)
	}
	if s.CPU.WeightNice != 19 {
		t.Errorf("cpu.weight.nice = %d, want 19", s.CPU.WeightNice)
	}
	if !s.CPU.Idle {
		t.Errorf("cpu.idle = false, want true")
	}
	if s.IO.PrioClass != ioPrioClassIdle {
		t.Errorf("io.prio.class = %q, want %q", s.IO.PrioClass, ioPrioClassIdle)
	}
	if s.IO.Weight != 1 {
		t.Errorf("io.weight = %d, want 1", s.IO.Weight)
	}
	if !s.Cgroup.Freeze {
		t.Errorf("cgroup.freeze = false, want true")
	}
	if s.Cgroup.Type != nodeTypeDomain {
		t.Errorf("cgroup.type = %q, want %q", s.Cgroup.Type, nodeTypeDomain)
	}
	if s.CPUSet.Partition != partitionMember {
		t.Errorf("cpuset.partition = %q, want %q", s.CPUSet.Partition, partitionMember)
	}
}
