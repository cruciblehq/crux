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
