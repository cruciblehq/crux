package blueprint

import (
	"errors"
	"strings"
	"testing"

	"github.com/cruciblehq/utils-go/crex"
)

func TestErrService(t *testing.T) {
	cause := crex.New("boom")

	err := errService("api", cause)
	if !errors.Is(err, ErrBuildPlan) {
		t.Fatalf("errService not wrapping ErrBuildPlan: %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("errService lost its cause: %v", err)
	}
	if !strings.Contains(err.Error(), "service api") {
		t.Fatalf("errService = %q, want it to name the service", err)
	}
}

func TestErrServiceAt(t *testing.T) {
	cause := crex.New("boom")

	err := errServiceAt("api", "/src/api", cause)
	if !errors.Is(err, ErrBuildPlan) {
		t.Fatalf("errServiceAt not wrapping ErrBuildPlan: %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("errServiceAt lost its cause: %v", err)
	}
	if !strings.Contains(err.Error(), "service api") || !strings.Contains(err.Error(), "/src/api") {
		t.Fatalf("errServiceAt = %q, want it to name the service and directory", err)
	}
}
