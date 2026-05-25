package spec

import "testing"

func TestFcapModeIsValid(t *testing.T) {
	if !FcapModeEffective.IsValid() {
		t.Error("effective should be valid")
	}
	if !FcapModeInheritable.IsValid() {
		t.Error("inheritable should be valid")
	}
}

func TestFcapModeIsValidRejectsUnknown(t *testing.T) {
	if FcapMode("other").IsValid() {
		t.Error("unknown mode should be invalid")
	}
	if FcapMode("").IsValid() {
		t.Error("empty mode should be invalid")
	}
}

func TestFcapCapabilitiesGrantEffectiveAdds(t *testing.T) {
	c := &FcapCapabilities{}
	changed := c.GrantEffective([]string{"cap_net_admin", "cap_net_bind_service"})
	if !changed {
		t.Error("expected changed=true")
	}
	if len(c.Permitted) != 2 {
		t.Fatalf("Permitted len = %d, want 2", len(c.Permitted))
	}
	if !c.Effective {
		t.Error("expected Effective=true")
	}
}

func TestFcapCapabilitiesGrantEffectiveDeduplicated(t *testing.T) {
	c := &FcapCapabilities{Permitted: []string{"cap_net_admin"}, Effective: true}
	changed := c.GrantEffective([]string{"cap_net_admin"})
	if changed {
		t.Error("expected changed=false for duplicate")
	}
	if len(c.Permitted) != 1 {
		t.Fatalf("Permitted len = %d, want 1", len(c.Permitted))
	}
}

func TestFcapCapabilitiesGrantInheritableAdds(t *testing.T) {
	c := &FcapCapabilities{}
	changed := c.GrantInheritable([]string{"cap_setuid"})
	if !changed {
		t.Error("expected changed=true")
	}
	if len(c.Inheritable) != 1 {
		t.Fatalf("Inheritable len = %d, want 1", len(c.Inheritable))
	}
}

func TestFcapCapabilitiesGrantInheritableDeduplicated(t *testing.T) {
	c := &FcapCapabilities{Inheritable: []string{"cap_setuid"}}
	changed := c.GrantInheritable([]string{"cap_setuid"})
	if changed {
		t.Error("expected changed=false for duplicate")
	}
}

func TestMergeStringSliceAddsNew(t *testing.T) {
	dst := []string{"a"}
	changed := mergeStringSlice(&dst, []string{"b", "c"})
	if !changed {
		t.Error("expected changed=true")
	}
	if len(dst) != 3 {
		t.Fatalf("len = %d, want 3", len(dst))
	}
}

func TestMergeStringSliceSkipsDuplicates(t *testing.T) {
	dst := []string{"a", "b"}
	changed := mergeStringSlice(&dst, []string{"a", "b"})
	if changed {
		t.Error("expected changed=false")
	}
	if len(dst) != 2 {
		t.Fatalf("len = %d, want 2", len(dst))
	}
}

