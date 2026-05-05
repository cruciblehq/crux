package manifest

import (
	"errors"
	"testing"
)

func TestManifestValidateOK(t *testing.T) {
	m := &Manifest{
		Resource: Resource{Type: TypeWidget, Name: "ns/w", Version: "1.0.0"},
		Config:   &Widget{Main: "index.js"},
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestManifestValidateAllResourceTypes(t *testing.T) {
	cases := []struct {
		typ ResourceType
		cfg any
	}{
		{TypeRuntime, &Runtime{Recipe: Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}}}},
		{TypeService, &Service{Recipe: Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}}, Entrypoint: []string{"/bin/run"}}},
		{TypeWidget, &Widget{Main: "index.js"}},
		{TypeTemplate, &Template{}},
		{TypeAffordance, &Affordance{}},
		{TypeBlueprint, &Blueprint{}},
	}
	for _, tc := range cases {
		m := &Manifest{
			Resource: Resource{Type: tc.typ, Name: "ns/x", Version: "1.0.0"},
			Config:   tc.cfg,
		}
		if err := m.Validate(); err != nil {
			t.Errorf("type %s: %v", tc.typ, err)
		}
	}
}

func TestManifestValidatePropagatesConfigError(t *testing.T) {
	m := &Manifest{
		Resource: Resource{Type: TypeWidget, Name: "ns/w", Version: "1.0.0"},
		Config:   &Widget{},
	}
	err := m.Validate()
	if !errors.Is(err, ErrMissingMain) {
		t.Fatalf("err = %v, want ErrMissingMain", err)
	}
}

func TestManifestValidatePropagatesResourceError(t *testing.T) {
	m := &Manifest{
		Resource: Resource{Type: TypeWidget, Version: "1.0.0"},
		Config:   &Widget{Main: "x"},
	}
	err := m.Validate()
	if !errors.Is(err, ErrMissingName) {
		t.Fatalf("err = %v, want ErrMissingName", err)
	}
}

func TestManifestValidateBadVersion(t *testing.T) {
	m := &Manifest{Version: 99}
	err := m.Validate()
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("err = %v, want ErrUnsupportedVersion", err)
	}
}

func TestManifestValidateMissingConfig(t *testing.T) {
	m := &Manifest{Resource: Resource{Type: TypeWidget, Name: "ns/w", Version: "1.0.0"}}
	err := m.Validate()
	if !errors.Is(err, ErrMissingConfig) {
		t.Fatalf("err = %v, want ErrMissingConfig", err)
	}
}

func TestManifestValidateConfigTypeMismatch(t *testing.T) {
	m := &Manifest{
		Resource: Resource{Type: TypeWidget, Name: "ns/w", Version: "1.0.0"},
		Config:   &Service{Entrypoint: []string{"x"}},
	}
	err := m.Validate()
	if !errors.Is(err, ErrConfigTypeMismatch) {
		t.Fatalf("err = %v, want ErrConfigTypeMismatch", err)
	}
}

func TestManifestValidateUnknownResourceType(t *testing.T) {
	m := &Manifest{
		Resource: Resource{Type: "bogus", Name: "ns/x", Version: "1.0.0"},
		Config:   &Widget{Main: "x"},
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestManifestEncodeFlatten(t *testing.T) {
	m := &Manifest{
		Resource: Resource{Type: TypeWidget, Name: "ns/w", Version: "1.0.0"},
		Config:   &Widget{Main: "index.js"},
	}
	out, err := m.Encode()
	if err != nil {
		t.Fatal(err)
	}
	mp := out.(map[string]any)
	if mp["main"] != "index.js" {
		t.Fatalf("main = %v", mp["main"])
	}
	if _, ok := mp["resource"]; !ok {
		t.Fatalf("missing resource key: %v", mp)
	}
}

func TestManifestDecode(t *testing.T) {
	raw := map[string]any{
		"version": 0,
		"resource": map[string]any{
			"type":    "widget",
			"name":    "ns/w",
			"version": "1.0.0",
		},
		"main": "index.js",
	}
	var m Manifest
	if err := m.Decode(raw); err != nil {
		t.Fatal(err)
	}
	w, ok := m.Config.(*Widget)
	if !ok {
		t.Fatalf("config = %T, want *Widget", m.Config)
	}
	if w.Main != "index.js" {
		t.Fatalf("main = %q", w.Main)
	}
}

func TestManifestDecodeRejectsNonMap(t *testing.T) {
	err := (&Manifest{}).Decode("nope")
	if !errors.Is(err, ErrDecodeFailed) {
		t.Fatalf("err = %v, want ErrDecodeFailed", err)
	}
}

func TestManifestDecodeUnknownResourceType(t *testing.T) {
	raw := map[string]any{
		"version": 0,
		"resource": map[string]any{
			"type":    "bogus",
			"name":    "ns/x",
			"version": "1.0.0",
		},
	}
	err := (&Manifest{}).Decode(raw)
	if !errors.Is(err, ErrInvalidResourceType) {
		t.Fatalf("err = %v, want ErrInvalidResourceType", err)
	}
}

func TestManifestDecodeAllResourceTypes(t *testing.T) {
	cases := []struct {
		typ    ResourceType
		extra  map[string]any
		assert func(t *testing.T, m *Manifest)
	}{
		{
			typ:   TypeRuntime,
			extra: map[string]any{"stages": []any{map[string]any{"steps": []any{map[string]any{"run": "x"}}}}},
			assert: func(t *testing.T, m *Manifest) {
				if _, ok := m.Config.(*Runtime); !ok {
					t.Fatalf("config = %T, want *Runtime", m.Config)
				}
			},
		},
		{
			typ: TypeService,
			extra: map[string]any{
				"stages":     []any{map[string]any{"steps": []any{map[string]any{"run": "x"}}}},
				"entrypoint": []any{"/bin/run"},
			},
			assert: func(t *testing.T, m *Manifest) {
				if _, ok := m.Config.(*Service); !ok {
					t.Fatalf("config = %T, want *Service", m.Config)
				}
			},
		},
		{
			typ:   TypeTemplate,
			extra: map[string]any{},
			assert: func(t *testing.T, m *Manifest) {
				if _, ok := m.Config.(*Template); !ok {
					t.Fatalf("config = %T, want *Template", m.Config)
				}
			},
		},
		{
			typ:   TypeAffordance,
			extra: map[string]any{"grants": []any{".cap effective net_admin"}},
			assert: func(t *testing.T, m *Manifest) {
				a, ok := m.Config.(*Affordance)
				if !ok {
					t.Fatalf("config = %T, want *Affordance", m.Config)
				}
				if len(a.Scopes) != 1 || len(a.Scopes[0].Grants) != 1 {
					t.Fatalf("scopes = %+v", a.Scopes)
				}
			},
		},
		{
			typ: TypeBlueprint,
			extra: map[string]any{
				"services": []any{map[string]any{"id": "svc", "ref": "ns/x"}},
				"gateway":  map[string]any{},
			},
			assert: func(t *testing.T, m *Manifest) {
				if _, ok := m.Config.(*Blueprint); !ok {
					t.Fatalf("config = %T, want *Blueprint", m.Config)
				}
			},
		},
	}
	for _, tc := range cases {
		raw := map[string]any{
			"version": 0,
			"resource": map[string]any{
				"type":    string(tc.typ),
				"name":    "ns/x",
				"version": "1.0.0",
			},
		}
		for k, v := range tc.extra {
			raw[k] = v
		}
		var m Manifest
		if err := m.Decode(raw); err != nil {
			t.Errorf("type %s: %v", tc.typ, err)
			continue
		}
		tc.assert(t, &m)
	}
}

func TestManifestEncodeAllResourceTypes(t *testing.T) {
	cases := []struct {
		typ ResourceType
		cfg any
	}{
		{TypeRuntime, &Runtime{Recipe: Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}}}},
		{TypeService, &Service{Recipe: Recipe{Stages: []Stage{{Steps: []Step{{Run: "x"}}}}}, Entrypoint: []string{"/bin/run"}}},
		{TypeTemplate, &Template{}},
		{TypeAffordance, &Affordance{Scopes: []GrantScope{{Grants: []Grant{{Source: ".cap effective net_admin"}}}}}},
		{TypeBlueprint, &Blueprint{Services: []Ref{{ID: "s", Target: "ns/x"}}}},
	}
	for _, tc := range cases {
		m := &Manifest{
			Resource: Resource{Type: tc.typ, Name: "ns/x", Version: "1.0.0"},
			Config:   tc.cfg,
		}
		out, err := m.Encode()
		if err != nil {
			t.Errorf("type %s: %v", tc.typ, err)
			continue
		}
		mp, ok := out.(map[string]any)
		if !ok {
			t.Errorf("type %s: out = %T", tc.typ, out)
			continue
		}
		if _, ok := mp["resource"]; !ok {
			t.Errorf("type %s: missing resource key", tc.typ)
		}
	}
}

func TestManifestEncodeRoundTrip(t *testing.T) {
	original := &Manifest{
		Resource: Resource{Type: TypeWidget, Name: "ns/w", Version: "1.0.0"},
		Config:   &Widget{Main: "index.js"},
	}
	out, err := original.Encode()
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err := m.Decode(out); err != nil {
		t.Fatal(err)
	}
	if m.Resource.Type != TypeWidget || m.Resource.Name != "ns/w" {
		t.Fatalf("resource = %+v", m.Resource)
	}
	w, ok := m.Config.(*Widget)
	if !ok || w.Main != "index.js" {
		t.Fatalf("config = %+v", m.Config)
	}
}
