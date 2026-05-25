package manifest

import (
	"errors"
	"testing"
)

func TestDeploymentValidateOK(t *testing.T) {
	d := &Deployment{Service: "svc", Container: "ctr", Compute: "c1", Network: "n1"}
	if err := d.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDeploymentValidateWithEnvironment(t *testing.T) {
	d := &Deployment{Service: "svc", Container: "ctr", Compute: "c1", Network: "n1", Environment: "env"}
	if err := d.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDeploymentValidateMissingService(t *testing.T) {
	d := &Deployment{Container: "ctr", Compute: "c1", Network: "n1"}
	err := d.Validate()
	if !errors.Is(err, ErrMissingDeploymentService) {
		t.Fatalf("err = %v, want ErrMissingDeploymentService", err)
	}
}

func TestDeploymentValidateMissingContainer(t *testing.T) {
	d := &Deployment{Service: "svc", Compute: "c1", Network: "n1"}
	err := d.Validate()
	if !errors.Is(err, ErrMissingDeploymentContainer) {
		t.Fatalf("err = %v, want ErrMissingDeploymentContainer", err)
	}
}

func TestDeploymentValidateMissingCompute(t *testing.T) {
	d := &Deployment{Service: "svc", Container: "ctr", Network: "n1"}
	err := d.Validate()
	if !errors.Is(err, ErrMissingDeploymentCompute) {
		t.Fatalf("err = %v, want ErrMissingDeploymentCompute", err)
	}
}

func TestDeploymentValidateMissingNetwork(t *testing.T) {
	d := &Deployment{Service: "svc", Container: "ctr", Compute: "c1"}
	err := d.Validate()
	if !errors.Is(err, ErrMissingDeploymentNetwork) {
		t.Fatalf("err = %v, want ErrMissingDeploymentNetwork", err)
	}
}
