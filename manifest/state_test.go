package manifest

import (
	"errors"
	"testing"
)

func TestStateValidateOK(t *testing.T) {
	s := &State{
		Version:     StateVersion,
		Deployments: []Deployment{{Service: "s1", Container: "s1", Compute: "c1", Network: "n1"}},
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStateValidateBadVersion(t *testing.T) {
	s := &State{Version: 99}
	err := s.Validate()
	if !errors.Is(err, ErrUnsupportedStateVersion) {
		t.Fatalf("err = %v, want ErrUnsupportedStateVersion", err)
	}
}

func TestStateValidateMissingDeploymentService(t *testing.T) {
	s := &State{
		Version:     StateVersion,
		Deployments: []Deployment{{Container: "c1", Compute: "c1", Network: "n1"}},
	}
	err := s.Validate()
	if !errors.Is(err, ErrMissingDeploymentService) {
		t.Fatalf("err = %v, want ErrMissingDeploymentService", err)
	}
}
