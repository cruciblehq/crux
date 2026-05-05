package manifest

import (
	"errors"
	"testing"
)

func TestStepValidateRun(t *testing.T) {
	if err := (&Step{Run: "echo hi"}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStepValidateCopy(t *testing.T) {
	if err := (&Step{Copy: "src dst"}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStepValidateModifierOnly(t *testing.T) {
	if err := (&Step{Workdir: "/app"}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStepValidateEmpty(t *testing.T) {
	err := (&Step{}).Validate()
	if !errors.Is(err, ErrEmptyStep) {
		t.Fatalf("err = %v, want ErrEmptyStep", err)
	}
}

func TestStepValidateRunAndCopyMutuallyExclusive(t *testing.T) {
	err := (&Step{Run: "x", Copy: "a b"}).Validate()
	if !errors.Is(err, ErrMutuallyExclusiveOps) {
		t.Fatalf("err = %v, want ErrMutuallyExclusiveOps", err)
	}
}

func TestStepValidateStepsRequirePlatform(t *testing.T) {
	s := &Step{Steps: []Step{{Run: "x"}}}
	err := s.Validate()
	if !errors.Is(err, ErrStepsWithoutPlatform) {
		t.Fatalf("err = %v, want ErrStepsWithoutPlatform", err)
	}
}

func TestStepValidatePlatformGroupNoOperation(t *testing.T) {
	s := &Step{Run: "x", Platform: "linux/amd64", Steps: []Step{{Run: "y"}}}
	err := s.Validate()
	if !errors.Is(err, ErrPlatformWithOperation) {
		t.Fatalf("err = %v, want ErrPlatformWithOperation", err)
	}
}

func TestStepValidatePlatformGroupOK(t *testing.T) {
	s := &Step{Platform: "linux/amd64", Steps: []Step{{Run: "x"}}}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStepValidateNestedPlatformGroupRejected(t *testing.T) {
	s := &Step{Platform: "linux/amd64", Steps: []Step{
		{Platform: "linux/arm64", Steps: []Step{{Run: "x"}}},
	}}
	err := s.Validate()
	if !errors.Is(err, ErrNestedPlatformGroup) {
		t.Fatalf("err = %v, want ErrNestedPlatformGroup", err)
	}
}

func TestStepValidateShellWithCopyRejected(t *testing.T) {
	err := (&Step{Copy: "a b", Shell: "/bin/bash"}).Validate()
	if !errors.Is(err, ErrShellWithCopy) {
		t.Fatalf("err = %v, want ErrShellWithCopy", err)
	}
}

func TestStepValidateEnvWithCopyRejected(t *testing.T) {
	err := (&Step{Copy: "a b", Env: map[string]string{"K": "V"}}).Validate()
	if !errors.Is(err, ErrEnvWithCopy) {
		t.Fatalf("err = %v, want ErrEnvWithCopy", err)
	}
}
