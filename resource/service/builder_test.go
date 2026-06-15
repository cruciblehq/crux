package service

import (
	"testing"

	"github.com/cruciblehq/crux/source"
)

func TestNewBuilder(t *testing.T) {
	src, err := source.NewSource("http://reg", "ns")
	if err != nil {
		t.Fatal(err)
	}

	b := NewBuilder(src, "/work")
	if b == nil {
		t.Fatal("NewBuilder returned nil")
	}
	if b.src != src {
		t.Errorf("src = %+v, want %+v", b.src, src)
	}
	if b.workdir != "/work" {
		t.Errorf("workdir = %q, want /work", b.workdir)
	}
}
