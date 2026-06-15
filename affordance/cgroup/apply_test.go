package cgroup

import (
	"errors"
	"testing"
)

func TestApplyKnobNestedScalar(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob("cpu.weight", "250"); err != nil {
		t.Fatal(err)
	}
	if s.CPU.Weight != 250 {
		t.Fatalf("cpu.weight = %d, want 250", s.CPU.Weight)
	}
}

func TestApplyKnobScalarSameValueConflicts(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob("cpu.weight", "250"); err != nil {
		t.Fatal(err)
	}
	err := s.applyKnob("cpu.weight", "250")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestApplyKnobScalarConflict(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob("cpu.weight", "100"); err != nil {
		t.Fatal(err)
	}
	err := s.applyKnob("cpu.weight", "200")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestApplyKnobUnknown(t *testing.T) {
	s := newSpec()
	err := s.applyKnob("foo.bar.baz", "1")
	if !errors.Is(err, ErrUnknownKnob) {
		t.Fatalf("err = %v, want ErrUnknownKnob", err)
	}
}

func TestApplyKnobIOWeightScalar(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob(ioWeightKnob, "500"); err != nil {
		t.Fatal(err)
	}
	if s.IO.Weight != 500 {
		t.Fatalf("io.weight = %d, want 500", s.IO.Weight)
	}
	if len(s.IO.WeightDevices) != 0 {
		t.Fatalf("WeightDevices populated unexpectedly: %v", s.IO.WeightDevices)
	}
}

func TestApplyKnobIOWeightPerDevice(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob(ioWeightKnob, "8:16 200"); err != nil {
		t.Fatal(err)
	}
	if got := s.IO.Weight; got == 200 {
		t.Fatalf("scalar Weight set unexpectedly to %d", got)
	}
	if len(s.IO.WeightDevices) != 1 {
		t.Fatalf("WeightDevices = %v, want 1 entry", s.IO.WeightDevices)
	}
	got := s.IO.WeightDevices[0]
	if got.Major != 8 || got.Minor != 16 || got.Weight != 200 {
		t.Fatalf("entry = %+v", got)
	}
}

func TestApplyKnobSubtreeControl(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob(subtreeControlKnob, "cpu memory"); err != nil {
		t.Fatal(err)
	}
	if len(s.Cgroup.SubtreeControl) != 2 {
		t.Fatalf("SubtreeControl = %v", s.Cgroup.SubtreeControl)
	}
}

func TestApplyKnobDmemRoutingByLeaf(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob(dmemMaxKnob, "gpu0 1024"); err != nil {
		t.Fatal(err)
	}
	if err := s.applyKnob(dmemMinKnob, "gpu0 512"); err != nil {
		t.Fatal(err)
	}
	if len(s.Dmem) != 1 {
		t.Fatalf("dmem = %v, want 1 entry merged by region", s.Dmem)
	}
	if s.Dmem[0].Max != 1024 || s.Dmem[0].Min != 512 {
		t.Fatalf("dmem[0] = %+v", s.Dmem[0])
	}
}

func TestApplyKnobPSIRoutingByBucket(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob(psiCPUKnob, "some 80 1000000"); err != nil {
		t.Fatal(err)
	}
	if err := s.applyKnob(psiMemoryKnob, "full 60 2000000"); err != nil {
		t.Fatal(err)
	}
	if len(s.PSI.CPU) != 1 || len(s.PSI.Memory) != 1 {
		t.Fatalf("PSI = %+v", s.PSI)
	}
}

func TestApplyKnobNamedScalarTypes(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob("cpuset.partition", "isolated"); err != nil {
		t.Fatal(err)
	}
	if s.CPUSet.Partition != partitionIsolated {
		t.Fatalf("partition = %q", s.CPUSet.Partition)
	}
	if err := s.applyKnob("cgroup.type", "threaded"); err != nil {
		t.Fatal(err)
	}
	if s.Cgroup.Type != nodeTypeThreaded {
		t.Fatalf("type = %q", s.Cgroup.Type)
	}
	if err := s.applyKnob("io.prio.class", "rt"); err != nil {
		t.Fatal(err)
	}
	if s.IO.PrioClass != ioPrioClassRT {
		t.Fatalf("prio.class = %q", s.IO.PrioClass)
	}
}

func TestApplyKnobIntAndFloatKinds(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob("cpu.weight.nice", "-5"); err != nil {
		t.Fatal(err)
	}
	if s.CPU.WeightNice != -5 {
		t.Fatalf("weight.nice = %d", s.CPU.WeightNice)
	}
	if err := s.applyKnob("cpu.uclamp.min", "0.25"); err != nil {
		t.Fatal(err)
	}
	if s.CPU.UclampMin != 0.25 {
		t.Fatalf("uclamp.min = %v", s.CPU.UclampMin)
	}
}

func TestApplyKnobIndexListSlice(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob("cpuset.cpus", "0-3,5"); err != nil {
		t.Fatal(err)
	}
	if len(s.CPUSet.CPUs) != 2 {
		t.Fatalf("cpus = %v", s.CPUSet.CPUs)
	}
}

func TestApplyKnobInvalidNumeric(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob("cpu.weight.nice", "not-a-number"); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
	if err := s.applyKnob("cpu.uclamp.min", "abc"); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v", err)
	}
}

func TestApplyKnobLazyInitSet(t *testing.T) {
	s := &spec{}
	if err := s.applyKnob("cpu.weight", "100"); err != nil {
		t.Fatal(err)
	}
	if s.CPU.Weight != 100 {
		t.Fatalf("cpu.weight = %d, want 100", s.CPU.Weight)
	}
}

func TestApplyKnobBoolFieldTrue(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob("cgroup.freeze", "1"); err != nil {
		t.Fatal(err)
	}
	if !s.Cgroup.Freeze {
		t.Fatal("freeze = false, want true")
	}
}

func TestApplyKnobBoolFieldInvalid(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob("cgroup.freeze", "yes"); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestApplyKnobIOWeightDeviceParseError(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob(ioWeightKnob, "8:16 abc"); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}
func TestApplyKnobIOWeightScalarParseError(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob(ioWeightKnob, "badvalue"); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}
func TestApplyKnobDeviceParseError(t *testing.T) {
	s := newSpec()
	if err := s.applyKnob(devicesKnob, "bad"); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}
