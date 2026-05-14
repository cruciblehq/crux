package manifest

import (
	"errors"
	"testing"
)

func TestStageValidateScratch(t *testing.T) {
	s := &Stage{Steps: []Step{{Run: "x"}}}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStageValidateWithFrom(t *testing.T) {
	s := &Stage{Name: "build", From: "ns/x", Steps: []Step{{Run: "x"}}}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStageValidateNumericName(t *testing.T) {
	s := &Stage{Name: "1"}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidStageName) {
		t.Fatalf("err = %v, want ErrInvalidStageName", err)
	}
}

func TestStageValidateInvalidName(t *testing.T) {
	s := &Stage{Name: "My Stage"}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidStageName) {
		t.Fatalf("err = %v, want ErrInvalidStageName", err)
	}
}

func TestStageValidateInvalidPlatform(t *testing.T) {
	s := &Stage{Platform: "Linux/AMD64", Steps: []Step{{Run: "x"}}}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidPlatform) {
		t.Fatalf("err = %v, want ErrInvalidPlatform", err)
	}
}

func TestStageValidateBadFrom(t *testing.T) {
	s := &Stage{Args: map[string]string{"k": "v"}}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error for args without from")
	}
}

func TestStageValidatePlatformBlocksStepPlatform(t *testing.T) {
	s := &Stage{
		Platform: "linux/amd64",
		Steps:    []Step{{Run: "x", Platform: "linux/arm64"}},
	}
	err := s.Validate()
	if !errors.Is(err, ErrPlatformInPlatformStage) {
		t.Fatalf("err = %v, want ErrPlatformInPlatformStage", err)
	}
}

func TestStageValidatePlatformBlocksNestedStepPlatform(t *testing.T) {
	s := &Stage{
		Platform: "linux/amd64",
		Steps: []Step{
			{Workdir: "/app", Steps: []Step{{Platform: "linux/arm64", Steps: []Step{{Run: "x"}}}}},
		},
	}
	err := s.Validate()
	if !errors.Is(err, ErrPlatformInPlatformStage) {
		t.Fatalf("err = %v, want ErrPlatformInPlatformStage", err)
	}
}

func TestStageValidatePropagatesGrantError(t *testing.T) {
	s := &Stage{Steps: []Step{{Run: "x"}}, Grants: []GrantScope{{Grants: []Grant{{Source: ""}}}}}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestStageValidateInlineGrants(t *testing.T) {
	s := &Stage{Steps: []Step{{Run: "x"}}, Grants: []GrantScope{{Grants: []Grant{{Source: ".seccomp read"}}}}}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStageValidatePlatformGrantScopeRejectedInPlatformStage(t *testing.T) {
	s := &Stage{
		Platform: "linux/amd64",
		Steps:    []Step{{Run: "x"}},
		Grants:   []GrantScope{{Platform: "linux/arm64", Grants: []Grant{{Source: ".seccomp read"}}}},
	}
	err := s.Validate()
	if !errors.Is(err, ErrGrantScopePlatformInPlatformStage) {
		t.Fatalf("err = %v, want ErrGrantScopePlatformInPlatformStage", err)
	}
}

func TestStageValidateStepError(t *testing.T) {
	s := &Stage{Steps: []Step{{Run: "x", Copy: "y"}}}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidStage) {
		t.Fatalf("err = %v, want ErrInvalidStage", err)
	}
}

func TestStageValidateGrantScopeError(t *testing.T) {
	s := &Stage{Grants: []GrantScope{{Grants: []Grant{{Source: ""}}}}}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidStage) {
		t.Fatalf("err = %v, want ErrInvalidStage", err)
	}
}

func TestStageEncodeUniversalGrants(t *testing.T) {
	s := &Stage{
		Steps:  []Step{{Run: "x"}},
		Grants: []GrantScope{{Grants: []Grant{{Source: ".seccomp openat"}}}},
	}
	raw, err := s.Encode()
	if err != nil {
		t.Fatal(err)
	}
	m, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("Encode() = %T, want map[string]any", raw)
	}
	list, ok := m["grants"].([]any)
	if !ok {
		t.Fatalf("grants = %T, want []any", m["grants"])
	}
	if len(list) != 1 {
		t.Fatalf("len(grants) = %d, want 1", len(list))
	}
	if list[0] != ".seccomp openat" {
		t.Fatalf("grants[0] = %v, want %q", list[0], ".seccomp openat")
	}
}

func TestStageEncodePlatformScopedGrants(t *testing.T) {
	s := &Stage{
		Steps: []Step{{Run: "x"}},
		Grants: []GrantScope{
			{Platform: "linux/amd64", Grants: []Grant{{Source: ".cap net_admin"}}},
		},
	}
	raw, err := s.Encode()
	if err != nil {
		t.Fatal(err)
	}
	m := raw.(map[string]any)
	list := m["grants"].([]any)
	if len(list) != 1 {
		t.Fatalf("len(grants) = %d, want 1", len(list))
	}
	group, ok := list[0].(map[string]any)
	if !ok {
		t.Fatalf("grants[0] = %T, want map", list[0])
	}
	if group["platform"] != "linux/amd64" {
		t.Fatalf("platform = %v, want linux/amd64", group["platform"])
	}
}

func TestStageDecodeGrants(t *testing.T) {
	src := map[string]any{
		"steps":  []any{map[string]any{"run": "x"}},
		"grants": []any{".seccomp openat", map[string]any{"platform": "linux/amd64", "grants": []any{".cap net_admin"}}},
	}
	var s Stage
	if err := s.Decode(src); err != nil {
		t.Fatal(err)
	}
	if len(s.Grants) != 2 {
		t.Fatalf("len(Grants) = %d, want 2", len(s.Grants))
	}
	if s.Grants[0].Platform != "" || len(s.Grants[0].Grants) != 1 {
		t.Fatalf("universal scope mismatch: %+v", s.Grants[0])
	}
	if s.Grants[1].Platform != "linux/amd64" || len(s.Grants[1].Grants) != 1 {
		t.Fatalf("platform scope mismatch: %+v", s.Grants[1])
	}
}

func TestStageEncodeDecodeRoundTrip(t *testing.T) {
	original := &Stage{
		Name:     "build",
		Platform: "linux/amd64",
		Steps:    []Step{{Run: "apt-get install -y curl"}},
		Grants: []GrantScope{
			{Grants: []Grant{{Source: ".seccomp openat"}}},
		},
	}
	enc, err := original.Encode()
	if err != nil {
		t.Fatal(err)
	}
	var decoded Stage
	if err := decoded.Decode(enc); err != nil {
		t.Fatal(err)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Platform != original.Platform {
		t.Errorf("Platform = %q, want %q", decoded.Platform, original.Platform)
	}
	if len(decoded.Steps) != 1 || decoded.Steps[0].Run != original.Steps[0].Run {
		t.Errorf("Steps mismatch: %+v", decoded.Steps)
	}
	if len(decoded.Grants) != 1 || decoded.Grants[0].Grants[0].Source != ".seccomp openat" {
		t.Errorf("Grants mismatch: %+v", decoded.Grants)
	}
}

func TestStageEncodeWithFrom(t *testing.T) {
	s := &Stage{
		From:  "ns/base",
		Steps: []Step{{Run: "x"}},
	}
	raw, err := s.Encode()
	if err != nil {
		t.Fatal(err)
	}
	m, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("Encode() = %T, want map[string]any", raw)
	}
	if m["from"] != "ns/base" {
		t.Fatalf("from = %v, want %q", m["from"], "ns/base")
	}
}

func TestStageDecodeInvalidType(t *testing.T) {
	var s Stage
	if err := s.Decode("not a map"); err == nil {
		t.Fatal("expected error for non-map input")
	}
}

func TestStageDecodeWithFrom(t *testing.T) {
	src := map[string]any{
		"from":  "ns/base",
		"steps": []any{map[string]any{"run": "x"}},
	}
	var s Stage
	if err := s.Decode(src); err != nil {
		t.Fatal(err)
	}
	if s.From != "ns/base" {
		t.Fatalf("From = %q, want %q", s.From, "ns/base")
	}
}
