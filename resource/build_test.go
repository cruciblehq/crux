package resource

import (
	"context"
	"errors"
	"testing"

	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/source"
)

func TestBuildUnsupportedType(t *testing.T) {
	var src source.Source
	m := manifest.Manifest{Resource: manifest.Resource{Type: manifest.ResourceType("bogus")}}
	if _, err := Build(context.Background(), m, src, "", "", ""); !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("Build(bogus) = %v, want ErrUnsupportedType", err)
	}
}
