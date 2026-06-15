package cgroup

import (
	"reflect"
	"testing"
)

func TestNewSpecInitialisesSeen(t *testing.T) {
	s := newSpec()
	if s.seen == nil {
		t.Fatal("seen map is nil")
	}
}

func TestSpecTopLevelFieldsAreTagged(t *testing.T) {
	t1 := reflect.TypeFor[spec]()
	for i := range t1.NumField() {
		f := t1.Field(i)
		if f.Name == "seen" { // tracker; not a knob target.
			continue
		}
		if f.Name == "PSI" { // routed by explicit intercept in applyKnob, not struct-tag walking.
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
		"HugeTLB": "hugetlb",
		"RDMA":    "rdma.max",
		"Misc":    "misc.max",
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

func TestSpecSeenFieldExcludedFromJSON(t *testing.T) {
	f, ok := reflect.TypeFor[spec]().FieldByName("seen")
	if !ok {
		t.Fatal("seen field missing")
	}
	if got := f.Tag.Get("json"); got != "-" {
		t.Fatalf("seen json tag = %q, want %q", got, "-")
	}
}
