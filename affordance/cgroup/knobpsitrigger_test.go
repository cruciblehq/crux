package cgroup

import (
	"errors"
	"testing"
)

func TestParsePSITrigger(t *testing.T) {
	got, err := parsePSITrigger("some 80 1000000")
	if err != nil {
		t.Fatal(err)
	}
	if got != (psiTrigger{Type: psiTypeSome, Threshold: 80, Window: 1000000}) {
		t.Fatalf("got %+v", got)
	}
}

func TestParsePSITriggerInvalid(t *testing.T) {
	for _, in := range []string{"", "some", "some 80", "bogus 80 1000", "some abc 1000", "some 80 abc"} {
		t.Run(in, func(t *testing.T) {
			_, err := parsePSITrigger(in)
			if !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestPSITriggerEqualByType(t *testing.T) {
	a := psiTrigger{Type: psiTypeSome, Threshold: 80, Window: 1000}
	if !a.equal(psiTrigger{Type: psiTypeSome, Threshold: 99, Window: 9999}) {
		t.Fatal("same type should be equal")
	}
	if a.equal(psiTrigger{Type: psiTypeFull}) {
		t.Fatal("different type should not be equal")
	}
}

func TestPSITriggerCheckRejectsDivergence(t *testing.T) {
	a := psiTrigger{Type: psiTypeSome, Threshold: 80, Window: 1000}
	b := psiTrigger{Type: psiTypeSome, Threshold: 90, Window: 1000}
	if err := a.check(b, psiCPUKnob); !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestPSITriggerCheckIdenticalNoConflict(t *testing.T) {
	a := psiTrigger{Type: psiTypeSome, Threshold: 80, Window: 1000}
	if err := a.check(a, psiCPUKnob); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestPSITriggerMergeNoOp(t *testing.T) {
	a := psiTrigger{Type: psiTypeSome, Threshold: 80, Window: 1000}
	if a.merge(a) {
		t.Fatal("merge reported change")
	}
}

func TestParsePSITriggerThresholdOverflow(t *testing.T) {
	_, err := parsePSITrigger("some 18446744073709551616 1000000")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestParsePSITriggerWindowOverflow(t *testing.T) {
	_, err := parsePSITrigger("some 80 18446744073709551616")
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}
