package resource

import (
	"context"
	"errors"
	"testing"

	"github.com/cruciblehq/crux/hub"
	"github.com/cruciblehq/spec/manifest"
	"github.com/cruciblehq/spec/registry"
)

func TestPushMissingPackage(t *testing.T) {
	src, err := hub.NewSource("http://reg", "ns")
	if err != nil {
		t.Fatal(err)
	}
	m := manifest.Manifest{
		Resource: manifest.Resource{Type: manifest.TypeWidget, Name: "ns/w", Version: "1.0.0"},
	}

	err = Push(context.Background(), src, m, "/nonexistent/pkg.tar.zst")
	if !errors.Is(err, registry.ErrFileSystemOperation) {
		t.Fatalf("Push missing package = %v, want registry.ErrFileSystemOperation", err)
	}
}
