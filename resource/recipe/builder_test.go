package recipe

import (
	"testing"

	"github.com/cruciblehq/spec/registry"
)

func TestNewBuilder(t *testing.T) {
	src, err := registry.NewSource("http://reg", "ns", "/cache")
	if err != nil {
		t.Fatal(err)
	}

	b := NewBuilder(src, "/work", nil)
	if b == nil {
		t.Fatal("NewBuilder returned nil")
	}
	if b.src != src {
		t.Errorf("src = %+v, want %+v", b.src, src)
	}
	if b.workdir != "/work" {
		t.Errorf("workdir = %q, want /work", b.workdir)
	}
	if b.agent != nil {
		t.Errorf("agent = %v, want nil", b.agent)
	}
}
