package kernel

import (
	"errors"
	"testing"
)

func TestSpecValidateEmptyFeature(t *testing.T) {
	s := &Spec{Features: []string{""}}
	if err := s.Validate(); !errors.Is(err, ErrInvalidSpec) {
		t.Fatalf("err = %v, want ErrInvalidSpec", err)
	}
}

func TestSpecValidateOK(t *testing.T) {
	s := &Spec{
		Features: []string{"NETFILTER"},
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSpecMergeFeatures(t *testing.T) {
	dst := Spec{Features: []string{"NETFILTER"}}
	src := Spec{Features: []string{"NETFILTER", "FUSE_FS"}}
	dst.Merge(src)
	if len(dst.Features) != 2 {
		t.Fatalf("Features = %v, want [NETFILTER FUSE_FS]", dst.Features)
	}
}
