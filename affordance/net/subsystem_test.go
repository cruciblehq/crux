package net

import (
	"errors"
	"testing"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
)

func nameArg(v string) agl.Arg { return agl.Arg{Type: agl.ArgName, Value: v} }
func intArg(v string) agl.Arg  { return agl.Arg{Type: agl.ArgInt, Value: v} }
func strArg(v string) agl.Arg  { return agl.Arg{Type: agl.ArgStrASCII, Value: v} }

func newSub() (*Subsystem, *Spec) {
	s := &Spec{}
	return New(s), s
}

func TestNameReturnsNet(t *testing.T) {
	sub, _ := newSub()
	if got := sub.Name(); got != subsystem.NameNet {
		t.Fatalf("Name() = %q, want %q", got, subsystem.NameNet)
	}
}

func TestBuildIngress(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("ingress"), nameArg("tcp"), intArg("8080")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(s.Ingress) != 1 {
		t.Fatalf("Ingress len = %d, want 1", len(s.Ingress))
	}
	r := s.Ingress[0]
	if r.Protocol != "tcp" || r.Port != 8080 {
		t.Fatalf("unexpected ingress rule: %+v", r)
	}
}

func TestBuildEgress(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("egress"), nameArg("tcp"), intArg("443"), nameArg("api.crucible.com")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(s.Egress) != 1 {
		t.Fatalf("Egress len = %d, want 1", len(s.Egress))
	}
	r := s.Egress[0]
	if r.Protocol != "tcp" || r.Port != 443 || r.Destination != "api.crucible.com" {
		t.Fatalf("unexpected egress rule: %+v", r)
	}
}

func TestBuildEgressCIDR(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("egress"), nameArg("tcp"), intArg("443"), strArg("10.0.0.0/8")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(s.Egress) != 1 || s.Egress[0].Destination != "10.0.0.0/8" {
		t.Fatalf("unexpected egress rule: %+v", s.Egress)
	}
}

func TestBuildRejectsWhere(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{
		Args:  []agl.Arg{nameArg("ingress"), nameArg("tcp"), intArg("80")},
		Where: &agl.CompareExpr{},
	}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsKwargs(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{
		Args:   []agl.Arg{nameArg("ingress"), nameArg("tcp"), intArg("80")},
		Kwargs: []agl.Kwarg{{Key: "k", Value: nameArg("v")}},
	}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsUnknownOp(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("listen"), nameArg("tcp"), intArg("80")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsUnknownProto(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("ingress"), nameArg("bogus"), intArg("80")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsPortlessIngressWithPort(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("ingress"), nameArg("icmp"), intArg("80")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildIngressPortless(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("ingress"), nameArg("icmp")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(s.Ingress) != 1 || s.Ingress[0].Protocol != "icmp" || s.Ingress[0].Port != 0 {
		t.Fatalf("unexpected ingress rule: %+v", s.Ingress)
	}
}

func TestBuildIngressWildcardIP(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("ingress"), nameArg("ip")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(s.Ingress) != 1 || s.Ingress[0].Protocol != "ip" {
		t.Fatalf("unexpected ingress rule: %+v", s.Ingress)
	}
}

func TestBuildIngressNumericProto(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("ingress"), nameArg("proto"), intArg("89")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(s.Ingress) != 1 || s.Ingress[0].Protocol != "89" || s.Ingress[0].Port != 0 {
		t.Fatalf("unexpected ingress rule: %+v", s.Ingress)
	}
}

func TestBuildEgressPortOmitted(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("egress"), nameArg("tcp"), nameArg("api.crucible.com")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	r := s.Egress[0]
	if r.Protocol != "tcp" || r.Port != 0 || r.Destination != "api.crucible.com" {
		t.Fatalf("unexpected egress rule: %+v", r)
	}
}

func TestBuildEgressPortless(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("egress"), nameArg("icmp"), strArg("*")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	r := s.Egress[0]
	if r.Protocol != "icmp" || r.Port != 0 || r.Destination != "*" {
		t.Fatalf("unexpected egress rule: %+v", r)
	}
}

func TestBuildEgressNumericProto(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("egress"), nameArg("proto"), intArg("89"), nameArg("mesh-router")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	r := s.Egress[0]
	if r.Protocol != "89" || r.Port != 0 || r.Destination != "mesh-router" {
		t.Fatalf("unexpected egress rule: %+v", r)
	}
}

func TestBuildIngressNumericProtoNormalizesToKeyword(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("ingress"), nameArg("proto"), intArg("6"), intArg("80")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	r := s.Ingress[0]
	if r.Protocol != "tcp" || r.Port != 80 {
		t.Fatalf("unexpected ingress rule: %+v", r)
	}
}

func TestBuildEgressNumericProtoNormalizesToKeyword(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("egress"), nameArg("proto"), intArg("17"), intArg("53"), nameArg("dns.crucible.com")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	r := s.Egress[0]
	if r.Protocol != "udp" || r.Port != 53 || r.Destination != "dns.crucible.com" {
		t.Fatalf("unexpected egress rule: %+v", r)
	}
}

func TestBuildRejectsEgressPortZero(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("egress"), nameArg("tcp"), intArg("0"), strArg("10.0.0.0/8")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildEgressAllNormalizesToIP(t *testing.T) {
	sub, s := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("egress"), nameArg("all"), strArg("*")}}
	if err := sub.Build(&g); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if s.Egress[0].Protocol != "ip" {
		t.Fatalf("Protocol = %q, want \"ip\"", s.Egress[0].Protocol)
	}
}

func TestBuildRejectsProtoNumberOutOfRange(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("egress"), nameArg("proto"), intArg("999"), nameArg("mesh")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsPortlessEgressWithPort(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("egress"), nameArg("icmp"), intArg("443"), strArg("*")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildIngressPortZero(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("ingress"), nameArg("tcp"), intArg("0")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildPortOutOfRange(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("ingress"), nameArg("tcp"), intArg("99999")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsWrongArgCountIngress(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("ingress"), nameArg("tcp")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsWrongArgCountEgress(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("egress"), nameArg("tcp"), intArg("443")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsNonNameProto(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("ingress"), intArg("6"), intArg("80")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsIntDestination(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("egress"), nameArg("tcp"), intArg("443"), intArg("123")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestIngressValidate(t *testing.T) {
	r := &IngressRule{Protocol: "tcp", Port: 80}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestIngressValidateUnknownProto(t *testing.T) {
	r := &IngressRule{Protocol: "bogus", Port: 80}
	if err := r.Validate(); !errors.Is(err, ErrInvalidIngressRule) {
		t.Fatalf("err = %v, want ErrInvalidIngressRule", err)
	}
}

func TestIngressValidatePortlessWithPort(t *testing.T) {
	r := &IngressRule{Protocol: "icmp", Port: 80}
	if err := r.Validate(); !errors.Is(err, ErrInvalidIngressRule) {
		t.Fatalf("err = %v, want ErrInvalidIngressRule", err)
	}
}

func TestIngressValidatePortless(t *testing.T) {
	r := &IngressRule{Protocol: "icmp"}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestIngressValidateZeroPort(t *testing.T) {
	r := &IngressRule{Protocol: "tcp", Port: 0}
	if err := r.Validate(); !errors.Is(err, ErrInvalidIngressRule) {
		t.Fatalf("err = %v, want ErrInvalidIngressRule", err)
	}
}

func TestEgressValidate(t *testing.T) {
	r := &EgressRule{Protocol: "udp", Port: 53, Destination: "8.8.8.8"}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestEgressValidateUnknownProto(t *testing.T) {
	r := &EgressRule{Protocol: "bogus", Destination: "example.com"}
	if err := r.Validate(); !errors.Is(err, ErrInvalidEgressRule) {
		t.Fatalf("err = %v, want ErrInvalidEgressRule", err)
	}
}

func TestEgressValidatePortless(t *testing.T) {
	r := &EgressRule{Protocol: "icmp", Destination: "example.com"}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestEgressValidatePortlessWithPort(t *testing.T) {
	r := &EgressRule{Protocol: "icmp", Port: 443, Destination: "example.com"}
	if err := r.Validate(); !errors.Is(err, ErrInvalidEgressRule) {
		t.Fatalf("err = %v, want ErrInvalidEgressRule", err)
	}
}

func TestEgressValidateEmptyDest(t *testing.T) {
	r := &EgressRule{Protocol: "tcp", Port: 443}
	if err := r.Validate(); !errors.Is(err, ErrInvalidEgressRule) {
		t.Fatalf("err = %v, want ErrInvalidEgressRule", err)
	}
}

func TestEgressValidateSubdomainWildcard(t *testing.T) {
	r := &EgressRule{Protocol: "tcp", Port: 443, Destination: "*.crucible.com"}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestEgressValidateUnrestricted(t *testing.T) {
	r := &EgressRule{Protocol: "tcp", Port: 443, Destination: "*"}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestEgressValidateRejectsGlob(t *testing.T) {
	for _, dest := range []string{"api.*", "*crucible*", "*.", "a.*.b"} {
		r := &EgressRule{Protocol: "tcp", Port: 443, Destination: dest}
		if err := r.Validate(); !errors.Is(err, ErrInvalidEgressRule) {
			t.Fatalf("dest %q: err = %v, want ErrInvalidEgressRule", dest, err)
		}
	}
}

func TestSpecValidate(t *testing.T) {
	p := &Spec{
		Ingress: []IngressRule{{Protocol: "tcp", Port: 80}},
		Egress:  []EgressRule{{Protocol: "tcp", Port: 443, Destination: "example.com"}},
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestSpecValidatePropagatesIngressError(t *testing.T) {
	p := &Spec{
		Ingress: []IngressRule{{Protocol: "bad", Port: 80}},
	}
	if err := p.Validate(); !errors.Is(err, ErrInvalidSpec) {
		t.Fatalf("err = %v, want ErrInvalidSpec", err)
	}
}

func TestSpecValidatePropagatesEgressError(t *testing.T) {
	p := &Spec{
		Egress: []EgressRule{{Protocol: "tcp"}},
	}
	if err := p.Validate(); !errors.Is(err, ErrInvalidSpec) {
		t.Fatalf("err = %v, want ErrInvalidSpec", err)
	}
}

func TestCheckRejectsNonNameOp(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{intArg("1"), nameArg("tcp"), intArg("80")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildPortNonInt(t *testing.T) {
	sub, _ := newSub()
	g := agl.Model{Args: []agl.Arg{nameArg("ingress"), nameArg("tcp"), nameArg("http")}}
	if err := sub.Build(&g); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}
