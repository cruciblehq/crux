package recipe

import (
	"testing"

	"github.com/cruciblehq/crux/hub"
)

func TestNewBuilder(t *testing.T) {
	src, err := hub.NewSource("http://reg", "ns")
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
	if b.client != nil {
		t.Errorf("client = %v, want nil", b.client)
	}
}
