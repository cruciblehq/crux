package cgroup

import (
	"errors"
	"testing"
)

func TestParsePSITypeValid(t *testing.T) {
	for _, in := range []psiType{psiTypeSome, psiTypeFull} {
		got, err := parsePSIType(string(in))
		if err != nil || got != in {
			t.Fatalf("got %q err %v", got, err)
		}
	}
}

func TestParsePSITypeInvalid(t *testing.T) {
	for _, in := range []string{"", "Some", "partial"} {
		_, err := parsePSIType(in)
		if !errors.Is(err, ErrInvalidGrant) {
			t.Fatalf("%q err = %v", in, err)
		}
	}
}

func TestMergePSIAppendsByType(t *testing.T) {
	var dst []psiTrigger
	src := []psiTrigger{
		{Type: psiTypeSome, Threshold: 80, Window: 1000},
		{Type: psiTypeFull, Threshold: 60, Window: 2000},
	}
	if err := mergePSI(&dst, src, psiCPUKnob); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 2 {
		t.Fatalf("dst = %v", dst)
	}
}

func TestMergePSIIdempotentForSameTrigger(t *testing.T) {
	dst := []psiTrigger{{Type: psiTypeSome, Threshold: 80, Window: 1000}}
	src := []psiTrigger{{Type: psiTypeSome, Threshold: 80, Window: 1000}}
	if err := mergePSI(&dst, src, psiCPUKnob); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 1 {
		t.Fatalf("dst = %v", dst)
	}
}

func TestMergePSIConflictOnDivergence(t *testing.T) {
	dst := []psiTrigger{{Type: psiTypeSome, Threshold: 80, Window: 1000}}
	src := []psiTrigger{{Type: psiTypeSome, Threshold: 90, Window: 1000}}
	err := mergePSI(&dst, src, psiCPUKnob)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestMergePSITriggersBucketSelection(t *testing.T) {
	cases := []struct {
		knob   string
		bucket func(*spec) []psiTrigger
	}{
		{psiCPUKnob, func(s *spec) []psiTrigger { return s.PSI.CPU }},
		{psiMemoryKnob, func(s *spec) []psiTrigger { return s.PSI.Memory }},
		{psiIOKnob, func(s *spec) []psiTrigger { return s.PSI.IO }},
		{psiIRQKnob, func(s *spec) []psiTrigger { return s.PSI.IRQ }},
	}
	for _, c := range cases {
		t.Run(c.knob, func(t *testing.T) {
			s := newSpec()
			tr := psiTrigger{Type: psiTypeSome, Threshold: 50, Window: 1000}
			if _, err := s.mergePSITriggers(c.knob, []psiTrigger{tr}); err != nil {
				t.Fatal(err)
			}
			if len(c.bucket(s)) != 1 {
				t.Fatalf("bucket %q empty after merge", c.knob)
			}
		})
	}
}

func TestMergePSITriggersUnknownKnob(t *testing.T) {
	s := newSpec()
	_, err := s.mergePSITriggers("psi.bogus", []psiTrigger{{Type: psiTypeSome, Threshold: 1, Window: 1}})
	if !errors.Is(err, ErrUnknownKnob) {
		t.Fatalf("err = %v", err)
	}
}
