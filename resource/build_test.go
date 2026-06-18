package resource

import (
	"context"
	"errors"
	"testing"

	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/registry"
)

func TestBuildUnsupportedType(t *testing.T) {
	var src registry.Source
	m := manifest.Manifest{Resource: manifest.Resource{Type: manifest.ResourceType("bogus")}}
	if _, err := Build(context.Background(), m, src, "", "", ""); !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("Build(bogus) = %v, want ErrUnsupportedType", err)
	}
}

func TestBuildMachineUnsupported(t *testing.T) {
	var src registry.Source
	m := manifest.Manifest{Resource: manifest.Resource{Type: manifest.TypeMachine}}
	if _, err := Build(context.Background(), m, src, "", "", ""); !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("Build(machine) = %v, want ErrUnsupportedType", err)
	}
}
