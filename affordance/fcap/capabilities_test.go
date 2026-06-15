package fcap

import (
	"errors"
	"testing"
)

func TestGrantEffectiveAddsAndSetsFlag(t *testing.T) {
	c := &Capabilities{}
	if !c.GrantEffective([]string{"CAP_NET_BIND_SERVICE"}) {
		t.Fatal("expected changed=true")
	}
	if !c.Effective {
		t.Fatal("expected Effective=true")
	}
	if len(c.Permitted) != 1 || c.Permitted[0] != "CAP_NET_BIND_SERVICE" {
		t.Fatalf("Permitted = %v", c.Permitted)
	}
}

func TestGrantEffectiveIdempotent(t *testing.T) {
	c := &Capabilities{}
	c.GrantEffective([]string{"CAP_CHOWN"})
	if c.GrantEffective([]string{"CAP_CHOWN"}) {
		t.Fatal("expected changed=false on second identical call")
	}
}

func TestGrantEffectiveAlreadySet(t *testing.T) {
	c := &Capabilities{Effective: true, Permitted: []string{"CAP_CHOWN"}}
	if c.GrantEffective([]string{"CAP_CHOWN"}) {
		t.Fatal("expected changed=false when no new state added")
	}
}

func TestGrantInheritableAdds(t *testing.T) {
	c := &Capabilities{}
	if !c.GrantInheritable([]string{"CAP_SYS_PTRACE"}) {
		t.Fatal("expected changed=true")
	}
	if len(c.Inheritable) != 1 {
		t.Fatalf("Inheritable = %v", c.Inheritable)
	}
}

func TestGrantInheritableIdempotent(t *testing.T) {
	c := &Capabilities{}
	c.GrantInheritable([]string{"CAP_SYS_PTRACE"})
	if c.GrantInheritable([]string{"CAP_SYS_PTRACE"}) {
		t.Fatal("expected changed=false on duplicate")
	}
}

func TestFcapCapabilitiesValidateEmpty(t *testing.T) {
	c := &Capabilities{}
	if err := c.Validate(); !errors.Is(err, ErrInvalidCapabilities) {
		t.Fatalf("err = %v, want ErrInvalidFcapCapabilities", err)
	}
}

func TestFcapCapabilitiesValidateEffectiveNoPermitted(t *testing.T) {
	c := &Capabilities{Effective: true, Inheritable: []string{"CAP_CHOWN"}}
	if err := c.Validate(); !errors.Is(err, ErrInvalidCapabilities) {
		t.Fatalf("err = %v, want ErrInvalidFcapCapabilities", err)
	}
}

func TestFcapCapabilitiesValidateInheritableOnly(t *testing.T) {
	c := &Capabilities{Inheritable: []string{"CAP_CHOWN"}}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFcapCapabilitiesValidateEffective(t *testing.T) {
	c := &Capabilities{Effective: true, Permitted: []string{"CAP_NET_BIND_SERVICE"}}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
