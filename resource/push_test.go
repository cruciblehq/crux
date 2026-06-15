package resource

import (
	"context"
	"errors"
	"testing"

	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/source"
)

func TestPushMissingPackage(t *testing.T) {
	src, err := source.NewSource("http://reg", "ns")
	if err != nil {
		t.Fatal(err)
	}
	m := manifest.Manifest{
		Resource: manifest.Resource{Type: manifest.TypeWidget, Name: "ns/w", Version: "1.0.0"},
	}

	err = Push(context.Background(), src, m, "/nonexistent/pkg.tar.zst")
	if !errors.Is(err, source.ErrFileSystemOperation) {
		t.Fatalf("Push missing package = %v, want source.ErrFileSystemOperation", err)
	}
}
