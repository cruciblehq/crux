package manifest

import (
	"errors"
	"testing"
)

func TestStageValidateScratch(t *testing.T) {
	s := &Stage{Steps: []Step{{Run: "x"}}}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStageValidateWithFrom(t *testing.T) {
	s := &Stage{Name: "build", From: &Ref{Target: "ns/x"}, Steps: []Step{{Run: "x"}}}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStageValidateNumericName(t *testing.T) {
	s := &Stage{Name: "1"}
	err := s.Validate()
	if !errors.Is(err, ErrNumericStageName) {
		t.Fatalf("err = %v, want ErrNumericStageName", err)
	}
}

func TestStageValidateBadFrom(t *testing.T) {
	s := &Stage{From: &Ref{}}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestStageValidatePlatformBlocksStepPlatform(t *testing.T) {
	s := &Stage{
		Platform: "linux/amd64",
		Steps:    []Step{{Run: "x", Platform: "linux/arm64"}},
	}
	err := s.Validate()
	if !errors.Is(err, ErrPlatformInPlatformStage) {
		t.Fatalf("err = %v, want ErrPlatformInPlatformStage", err)
	}
}

func TestStageValidatePlatformBlocksNestedStepPlatform(t *testing.T) {
	s := &Stage{
		Platform: "linux/amd64",
		Steps: []Step{
			{Workdir: "/app", Steps: []Step{{Platform: "linux/arm64", Steps: []Step{{Run: "x"}}}}},
		},
	}
	err := s.Validate()
	if !errors.Is(err, ErrPlatformInPlatformStage) {
		t.Fatalf("err = %v, want ErrPlatformInPlatformStage", err)
	}
}

func TestStageValidatePropagatesGrantError(t *testing.T) {
	s := &Stage{Steps: []Step{{Run: "x"}}, Grants: []Grant{{Source: ""}}}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestStageValidateInlineGrants(t *testing.T) {
	s := &Stage{Steps: []Step{{Run: "x"}}, Grants: []Grant{{Source: ".seccomp read"}}}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}
