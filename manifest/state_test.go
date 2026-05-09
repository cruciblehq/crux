package manifest

import (
	"errors"
	"testing"
	"time"
)

func TestStateValidateOK(t *testing.T) {
	s := &State{
		Version:    StateVersion,
		Deployment: Deployment{DeployedAt: time.Now()},
		Services:   []Ref{{ID: "s1", Target: "ns/x"}},
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStateValidateBadVersion(t *testing.T) {
	s := &State{Version: 99, Deployment: Deployment{DeployedAt: time.Now()}}
	err := s.Validate()
	if !errors.Is(err, ErrUnsupportedStateVersion) {
		t.Fatalf("err = %v, want ErrUnsupportedStateVersion", err)
	}
}

func TestStateValidateMissingDeployedAt(t *testing.T) {
	err := (&State{Version: StateVersion}).Validate()
	if !errors.Is(err, ErrMissingDeployedAt) {
		t.Fatalf("err = %v, want ErrMissingDeployedAt", err)
	}
}

func TestStateValidateServiceMissingID(t *testing.T) {
	s := &State{
		Version:    StateVersion,
		Deployment: Deployment{DeployedAt: time.Now()},
		Services:   []Ref{{Target: "ns/x"}},
	}
	err := s.Validate()
	if !errors.Is(err, ErrMissingServiceID) {
		t.Fatalf("err = %v, want ErrMissingServiceID", err)
	}
}
