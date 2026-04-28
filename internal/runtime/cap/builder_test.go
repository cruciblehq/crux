package cap

import (
	"slices"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func TestComposeNilBuilderIsNoOp(t *testing.T) {
	b := NewBuilder()
	target := &specs.LinuxCapabilities{}
	if err := b.Compose(target); err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if len(target.Effective) != 0 {
		t.Fatal("nil builder produced capabilities")
	}
}

func TestComposeFullMode(t *testing.T) {
	b := NewBuilder()
	b.apply("net_admin", ModeFull)
	target := &specs.LinuxCapabilities{}
	if err := b.Compose(target); err != nil {
		t.Fatalf("Compose: %v", err)
	}
	want := []string{"CAP_NET_ADMIN"}
	for name, got := range map[string][]string{
		"effective":   target.Effective,
		"permitted":   target.Permitted,
		"inheritable": target.Inheritable,
		"bounding":    target.Bounding,
		"ambient":     target.Ambient,
	} {
		if !slices.Equal(got, want) {
			t.Errorf("%s = %v, want %v", name, got, want)
		}
	}
}

func TestComposeDedupesAgainstBaseline(t *testing.T) {
	b := NewBuilder()
	b.apply("chown", ModeEffective)
	target := &specs.LinuxCapabilities{Effective: []string{"CAP_CHOWN"}}
	if err := b.Compose(target); err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if !slices.Equal(target.Effective, []string{"CAP_CHOWN"}) {
		t.Fatalf("Effective = %v, want [CAP_CHOWN]", target.Effective)
	}
}

func TestSpecReturnsEmptyWhenNoGrants(t *testing.T) {
	b := NewBuilder()
	got := b.Spec()
	if got == nil {
		t.Fatal("Model() = nil, want empty value")
	}
	if len(got.Effective) != 0 || len(got.Permitted) != 0 ||
		len(got.Inheritable) != 0 || len(got.Bounding) != 0 || len(got.Ambient) != 0 {
		t.Fatalf("Model() = %+v, want all empty", got)
	}
}

func TestSpecClonesAccumulatedState(t *testing.T) {
	b := NewBuilder()
	b.apply("net_admin", ModeBound)
	got := b.Spec()
	if got == nil {
		t.Fatal("Model() returned nil")
	}
	if !slices.Equal(got.Bounding, []string{"CAP_NET_ADMIN"}) {
		t.Fatalf("Bounding = %v", got.Bounding)
	}
	got.Bounding[0] = "CAP_MUTATED"
	if b.caps.Bounding[0] != "CAP_NET_ADMIN" {
		t.Fatal("Model() did not clone")
	}
}

func TestMergeUnionsSets(t *testing.T) {
	b := NewBuilder()
	b.apply("chown", ModeBound)
	other := &specs.LinuxCapabilities{
		Bounding:  []string{"CAP_CHOWN", "CAP_NET_ADMIN"},
		Effective: []string{"CAP_KILL"},
	}
	if err := b.Merge(other); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if !slices.Equal(b.caps.Bounding, []string{"CAP_CHOWN", "CAP_NET_ADMIN"}) {
		t.Fatalf("Bounding = %v", b.caps.Bounding)
	}
	if !slices.Equal(b.caps.Effective, []string{"CAP_KILL"}) {
		t.Fatalf("Effective = %v", b.caps.Effective)
	}
}
