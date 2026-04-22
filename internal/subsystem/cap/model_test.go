package cap

import (
	"errors"
	"testing"
)

func TestCloneModelReturnsNilForNilOrEmpty(t *testing.T) {
	if cloneModel(nil) != nil {
		t.Fatal("cloneModel(nil) != nil")
	}
	if cloneModel(&Model{}) != nil {
		t.Fatal("cloneModel(empty) != nil")
	}
}

func TestCloneModelReturnsDeepCopy(t *testing.T) {
	model := &Model{
		Effective:   []string{"net_admin"},
		Permitted:   []string{"net_admin"},
		Inheritable: []string{"net_admin"},
		Bounding:    []string{"net_admin"},
		Ambient:     []string{"net_admin"},
	}
	clone := cloneModel(model)
	clone.Effective[0] = "sys_admin"
	clone.Permitted[0] = "sys_admin"
	clone.Inheritable[0] = "sys_admin"
	clone.Bounding[0] = "sys_admin"
	clone.Ambient[0] = "sys_admin"

	if model.Effective[0] != "net_admin" {
		t.Fatalf("effective = %#v", model.Effective)
	}
	if model.Permitted[0] != "net_admin" {
		t.Fatalf("permitted = %#v", model.Permitted)
	}
	if model.Inheritable[0] != "net_admin" {
		t.Fatalf("inheritable = %#v", model.Inheritable)
	}
	if model.Bounding[0] != "net_admin" {
		t.Fatalf("bounding = %#v", model.Bounding)
	}
	if model.Ambient[0] != "net_admin" {
		t.Fatalf("ambient = %#v", model.Ambient)
	}
}

func TestModelApplyByMode(t *testing.T) {
	tests := []struct {
		name        string
		rule        string
		effective   []string
		permitted   []string
		inheritable []string
		bounding    []string
		ambient     []string
	}{
		{name: "full", rule: "net_admin", effective: []string{"net_admin"}, permitted: []string{"net_admin"}, inheritable: []string{"net_admin"}, bounding: []string{"net_admin"}, ambient: []string{"net_admin"}},
		{name: "effective", rule: "effective net_admin", effective: []string{"net_admin"}, permitted: []string{"net_admin"}, bounding: []string{"net_admin"}},
		{name: "inheritable", rule: "inheritable net_admin", permitted: []string{"net_admin"}, inheritable: []string{"net_admin"}, bounding: []string{"net_admin"}, ambient: []string{"net_admin"}},
		{name: "permitted", rule: "permitted net_admin", permitted: []string{"net_admin"}, bounding: []string{"net_admin"}},
		{name: "bound", rule: "bound net_admin", bounding: []string{"net_admin"}},
	}

	for _, test := range tests {
		model := NewModel()
		delta, err := Parse(test.rule)
		if err != nil {
			t.Fatalf("%s parse: %v", test.name, err)
		}
		model.Merge(delta)
		if clone := cloneModel(model); clone == nil {
			t.Fatalf("%s: model = nil after merge", test.name)
		}
		assertSlice(t, test.name+" effective", model.Effective, test.effective)
		assertSlice(t, test.name+" permitted", model.Permitted, test.permitted)
		assertSlice(t, test.name+" inheritable", model.Inheritable, test.inheritable)
		assertSlice(t, test.name+" bounding", model.Bounding, test.bounding)
		assertSlice(t, test.name+" ambient", model.Ambient, test.ambient)

		model.Merge(delta)
		assertSlice(t, test.name+" second effective", model.Effective, test.effective)
		assertSlice(t, test.name+" second permitted", model.Permitted, test.permitted)
		assertSlice(t, test.name+" second inheritable", model.Inheritable, test.inheritable)
		assertSlice(t, test.name+" second bounding", model.Bounding, test.bounding)
		assertSlice(t, test.name+" second ambient", model.Ambient, test.ambient)
	}
}

func TestModelApplyUnknownModeReturnsError(t *testing.T) {
	model := NewModel()
	_, err := Parse("bogus net_admin")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
	if cloneModel(model) != nil {
		t.Fatal("model changed on invalid mode")
	}
}

func TestModelMergeDeduplicatesBySet(t *testing.T) {
	model := NewModel()
	model.Merge(&Model{
		Effective:   []string{"net_admin"},
		Permitted:   []string{"net_admin"},
		Inheritable: []string{"net_admin"},
		Bounding:    []string{"net_admin"},
		Ambient:     []string{"net_admin"},
	})
	model.Merge(&Model{
		Effective:   []string{"net_admin", "sys_admin"},
		Permitted:   []string{"net_admin", "sys_admin"},
		Inheritable: []string{"net_admin", "sys_admin"},
		Bounding:    []string{"net_admin", "sys_admin"},
		Ambient:     []string{"net_admin", "sys_admin"},
	})

	assertSlice(t, "effective", model.Effective, []string{"net_admin", "sys_admin"})
	assertSlice(t, "permitted", model.Permitted, []string{"net_admin", "sys_admin"})
	assertSlice(t, "inheritable", model.Inheritable, []string{"net_admin", "sys_admin"})
	assertSlice(t, "bounding", model.Bounding, []string{"net_admin", "sys_admin"})
	assertSlice(t, "ambient", model.Ambient, []string{"net_admin", "sys_admin"})
}

func assertSlice(t *testing.T, name string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s len = %d, want %d: %#v", name, len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %q, want %q", name, i, got[i], want[i])
		}
	}
}
