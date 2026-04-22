package cap

import (
	"errors"
	"testing"
)

func TestNewBuilderStartsEmpty(t *testing.T) {
	if NewBuilder().Model() != nil {
		t.Fatal("model != nil")
	}
}

func TestBuilderBuildAccumulatesCapabilityModel(t *testing.T) {
	b := NewBuilder()
	if err := b.Build("effective net_admin"); err != nil {
		t.Fatal(err)
	}

	s := b.Model()
	if s == nil {
		t.Fatal("model = nil")
	}
	assertSlice(t, "effective", s.Effective, []string{"net_admin"})
	assertSlice(t, "permitted", s.Permitted, []string{"net_admin"})
	assertSlice(t, "bounding", s.Bounding, []string{"net_admin"})
	assertSlice(t, "inheritable", s.Inheritable, nil)
	assertSlice(t, "ambient", s.Ambient, nil)
}

func TestBuilderBuildErrorLeavesModelEmpty(t *testing.T) {
	b := NewBuilder()
	err := b.Build("effective not_a_cap")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
	if b.Model() != nil {
		t.Fatal("model != nil")
	}
}

func TestBuilderModelReturnsIsolatedSnapshot(t *testing.T) {
	b := NewBuilder()
	if err := b.Build("effective net_admin"); err != nil {
		t.Fatal(err)
	}

	s := b.Model()
	s.Effective[0] = "sys_admin"
	s.Permitted[0] = "sys_admin"
	s.Bounding[0] = "sys_admin"

	again := b.Model()
	assertSlice(t, "effective", again.Effective, []string{"net_admin"})
	assertSlice(t, "permitted", again.Permitted, []string{"net_admin"})
	assertSlice(t, "bounding", again.Bounding, []string{"net_admin"})
}

func TestNewBuilderWithModelClonesInput(t *testing.T) {
	seed := &Model{
		Effective: []string{"net_admin"},
		Permitted: []string{"net_admin"},
		Bounding:  []string{"net_admin"},
	}
	b := NewBuilderWithModel(seed)
	seed.Effective[0] = "sys_admin"
	seed.Permitted[0] = "sys_admin"
	seed.Bounding[0] = "sys_admin"

	s := b.Model()
	assertSlice(t, "effective", s.Effective, []string{"net_admin"})
	assertSlice(t, "permitted", s.Permitted, []string{"net_admin"})
	assertSlice(t, "bounding", s.Bounding, []string{"net_admin"})
}

func TestBuilderMergeNilIsNoOp(t *testing.T) {
	b := NewBuilder()
	if err := b.Merge(nil); err != nil {
		t.Fatal(err)
	}
	if b.Model() != nil {
		t.Fatal("model != nil")
	}
}

func TestBuilderMergeClonesInputWhenEmpty(t *testing.T) {
	other := &Model{
		Effective: []string{"net_admin"},
		Permitted: []string{"net_admin"},
		Bounding:  []string{"net_admin"},
	}
	b := NewBuilder()
	if err := b.Merge(other); err != nil {
		t.Fatal(err)
	}
	other.Effective[0] = "sys_admin"
	other.Permitted[0] = "sys_admin"
	other.Bounding[0] = "sys_admin"

	s := b.Model()
	assertSlice(t, "effective", s.Effective, []string{"net_admin"})
	assertSlice(t, "permitted", s.Permitted, []string{"net_admin"})
	assertSlice(t, "bounding", s.Bounding, []string{"net_admin"})
}

func TestBuilderMergeMergesSetsIdempotently(t *testing.T) {
	b := NewBuilder()
	if err := b.Merge(&Model{
		Effective: []string{"net_admin"},
		Permitted: []string{"net_admin"},
		Bounding:  []string{"net_admin"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := b.Merge(&Model{
		Effective: []string{"net_admin", "sys_admin"},
		Permitted: []string{"net_admin", "sys_admin"},
		Bounding:  []string{"net_admin", "sys_admin"},
	}); err != nil {
		t.Fatal(err)
	}

	s := b.Model()
	assertSlice(t, "effective", s.Effective, []string{"net_admin", "sys_admin"})
	assertSlice(t, "permitted", s.Permitted, []string{"net_admin", "sys_admin"})
	assertSlice(t, "bounding", s.Bounding, []string{"net_admin", "sys_admin"})
}
