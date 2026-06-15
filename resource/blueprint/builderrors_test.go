package blueprint

import (
	"errors"
	"strings"
	"testing"

	"github.com/cruciblehq/crux/crex"
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

func TestErrServiceRuntime(t *testing.T) {
	cause := crex.New("boom")

	err := errServiceRuntime("api", cause)
	if !errors.Is(err, ErrBuildPlan) {
		t.Fatalf("errServiceRuntime not wrapping ErrBuildPlan: %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("errServiceRuntime lost its cause: %v", err)
	}
	if !strings.Contains(err.Error(), "runtime") {
		t.Fatalf("errServiceRuntime = %q, want it to mention runtime", err)
	}
}
