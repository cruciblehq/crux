package reference

import (
	"errors"
	"testing"
)

func TestReferenceParser_Parse_NameAndVersion(t *testing.T) {
	p := &referenceParser{
		tokens:  []string{"widget", "1.0.0"},
		options: &IdentifierOptions{DefaultRegistry: "https://registry.test"},
	}

	ref, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if ref.Type() != "template" {
		t.Errorf("expected type %q, got %q", "template", ref.Type())
	}
	if ref.Name() != "widget" {
		t.Errorf("expected name %q, got %q", "widget", ref.Name())
	}
	if ref.version == nil {
		t.Fatal("expected version, got nil")
	}
}

func TestReferenceParser_Parse_NamespaceNameAndVersion(t *testing.T) {
	p := &referenceParser{
		tokens:  []string{"namespace/name", "1.0.0"},
		options: &IdentifierOptions{DefaultRegistry: "https://registry.test"},
	}

	ref, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if ref.Namespace() != "namespace" {
		t.Errorf("expected namespace %q, got %q", "namespace", ref.Namespace())
	}
	if ref.Name() != "name" {
		t.Errorf("expected name %q, got %q", "name", ref.Name())
	}
}

func TestReferenceParser_Parse_WithExplicitType(t *testing.T) {
	p := &referenceParser{
		tokens:  []string{"template", "namespace/name", "1.0.0"},
		options: &IdentifierOptions{DefaultRegistry: "https://registry.test"},
	}

	ref, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if ref.Type() != "template" {
		t.Errorf("expected type %q, got %q", "template", ref.Type())
	}
	if ref.Namespace() != "namespace" {
		t.Errorf("expected namespace %q, got %q", "namespace", ref.Namespace())
	}
}

func TestReferenceParser_Parse_FullURIAndVersion(t *testing.T) {
	p := &referenceParser{
		tokens:  []string{"https://myregistry.com/path/to/resource", "1.0.0"},
		options: &IdentifierOptions{DefaultRegistry: "https://registry.test"},
	}

	ref, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if ref.Registry() != "https://myregistry.com" {
		t.Errorf("expected registry %q, got %q", "https://myregistry.com", ref.Registry())
	}
	if ref.Path() != "path/to/resource" {
		t.Errorf("expected path %q, got %q", "path/to/resource", ref.Path())
	}
}

func TestReferenceParser_Parse_WithChannel(t *testing.T) {
	p := &referenceParser{
		tokens:  []string{"namespace/name", ":stable"},
		options: &IdentifierOptions{DefaultRegistry: "https://registry.test"},
	}

	ref, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if ref.channel == nil {
		t.Fatal("expected channel, got nil")
	}
	if *ref.channel != "stable" {
		t.Errorf("expected channel %q, got %q", "stable", *ref.channel)
	}
	if ref.version != nil {
		t.Error("expected version to be nil")
	}
}

func TestReferenceParser_Parse_WithVersionRange(t *testing.T) {
	p := &referenceParser{
		tokens:  []string{"namespace/name", ">=1.0.0", "<2.0.0"},
		options: &IdentifierOptions{DefaultRegistry: "https://registry.test"},
	}

	ref, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if ref.version == nil {
		t.Fatal("expected version, got nil")
	}
}

func TestReferenceParser_Parse_WithDigest(t *testing.T) {
	p := &referenceParser{
		tokens:  []string{"namespace/name", "1.0.0", "sha256:abcd1234"},
		options: &IdentifierOptions{DefaultRegistry: "https://registry.test"},
	}

	ref, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if ref.digest == nil {
		t.Fatal("expected digest, got nil")
	}
}

func TestReferenceParser_Parse_ChannelWithDigest(t *testing.T) {
	p := &referenceParser{
		tokens:  []string{"namespace/name", ":stable", "sha256:abcd1234"},
		options: &IdentifierOptions{DefaultRegistry: "https://registry.test"},
	}

	ref, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if ref.channel == nil {
		t.Fatal("expected channel, got nil")
	}
	if ref.digest == nil {
		t.Fatal("expected digest, got nil")
	}
}

func TestReferenceParser_Parse_WithOptions(t *testing.T) {
	p := &referenceParser{
		tokens: []string{"widget", "1.0.0"},
		options: &IdentifierOptions{
			DefaultRegistry:  "https://custom.registry.io",
			DefaultNamespace: "myteam",
		},
	}

	ref, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if ref.Registry() != "https://custom.registry.io" {
		t.Errorf("expected registry %q, got %q", "https://custom.registry.io", ref.Registry())
	}
	if ref.Namespace() != "myteam" {
		t.Errorf("expected namespace %q, got %q", "myteam", ref.Namespace())
	}
}

func TestReferenceParser_Parse_EmptyTokens(t *testing.T) {
	p := &referenceParser{
		tokens:  []string{},
		options: &IdentifierOptions{DefaultRegistry: "https://registry.test"},
	}

	_, err := p.parse("template")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidReference) {
		t.Errorf("expected ErrInvalidReference, got %v", err)
	}
}

func TestReferenceParser_Parse_MissingVersion(t *testing.T) {
	p := &referenceParser{
		tokens:  []string{"namespace/name"},
		options: &IdentifierOptions{DefaultRegistry: "https://registry.test"},
	}

	_, err := p.parse("template")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidReference) {
		t.Errorf("expected ErrInvalidReference, got %v", err)
	}
}

func TestReferenceParser_Parse_TypeMismatch(t *testing.T) {
	p := &referenceParser{
		tokens:  []string{"plugin", "namespace/name", "1.0.0"},
		options: &IdentifierOptions{DefaultRegistry: "https://registry.test"},
	}

	_, err := p.parse("template")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrTypeMismatch) {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

func TestReferenceParser_Parse_UnexpectedToken(t *testing.T) {
	p := &referenceParser{
		tokens:  []string{"namespace/name", "1.0.0", "extra"},
		options: &IdentifierOptions{DefaultRegistry: "https://registry.test"},
	}

	_, err := p.parse("template")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidReference) {
		t.Errorf("expected ErrInvalidReference, got %v", err)
	}
}

func TestReferenceParser_FindIdentifierEnd_WithVersion(t *testing.T) {
	p := &referenceParser{
		tokens: []string{"namespace/name", "1.0.0"},
	}

	end := p.findIdentifierEnd()
	if end != 1 {
		t.Errorf("expected end 1, got %d", end)
	}
}

func TestReferenceParser_FindIdentifierEnd_WithChannel(t *testing.T) {
	p := &referenceParser{
		tokens: []string{"namespace/name", ":stable"},
	}

	end := p.findIdentifierEnd()
	if end != 1 {
		t.Errorf("expected end 1, got %d", end)
	}
}

func TestReferenceParser_FindIdentifierEnd_WithDigest(t *testing.T) {
	p := &referenceParser{
		tokens: []string{"namespace/name", "sha256:abcd1234"},
	}

	end := p.findIdentifierEnd()
	if end != 1 {
		t.Errorf("expected end 1, got %d", end)
	}
}

func TestReferenceParser_FindIdentifierEnd_WithTypeAndVersion(t *testing.T) {
	p := &referenceParser{
		tokens: []string{"template", "namespace/name", "1.0.0"},
	}

	end := p.findIdentifierEnd()
	if end != 2 {
		t.Errorf("expected end 2, got %d", end)
	}
}

func TestReferenceParser_FindIdentifierEnd_NoVersionOrChannel(t *testing.T) {
	p := &referenceParser{
		tokens: []string{"namespace/name"},
	}

	end := p.findIdentifierEnd()
	if end != 1 {
		t.Errorf("expected end 1, got %d", end)
	}
}

func TestReferenceParser_Peek(t *testing.T) {
	p := &referenceParser{
		tokens: []string{"a", "b", "c"},
	}

	tok, ok := p.peek()
	if !ok || tok != "a" {
		t.Errorf("expected %q, got %q", "a", tok)
	}

	tok, ok = p.peek()
	if !ok || tok != "a" {
		t.Errorf("expected %q, got %q", "a", tok)
	}
}

func TestReferenceParser_Peek_Empty(t *testing.T) {
	p := &referenceParser{
		tokens: []string{},
	}

	_, ok := p.peek()
	if ok {
		t.Error("expected ok to be false")
	}
}

func TestReferenceParser_Next(t *testing.T) {
	p := &referenceParser{
		tokens: []string{"a", "b", "c"},
	}

	tok, ok := p.next()
	if !ok || tok != "a" {
		t.Errorf("expected %q, got %q", "a", tok)
	}

	tok, ok = p.next()
	if !ok || tok != "b" {
		t.Errorf("expected %q, got %q", "b", tok)
	}

	tok, ok = p.next()
	if !ok || tok != "c" {
		t.Errorf("expected %q, got %q", "c", tok)
	}

	_, ok = p.next()
	if ok {
		t.Error("expected ok to be false")
	}
}

func TestReferenceParser_Remaining(t *testing.T) {
	p := &referenceParser{
		tokens: []string{"a", "b", "c"},
	}

	if p.remaining() != 3 {
		t.Errorf("expected remaining 3, got %d", p.remaining())
	}

	p.next()
	if p.remaining() != 2 {
		t.Errorf("expected remaining 2, got %d", p.remaining())
	}
}

func TestReferenceParser_ParseVersionOrChannel_Version(t *testing.T) {
	p := &referenceParser{
		tokens: []string{"1.0.0"},
		pos:    0,
	}

	ref := &Reference{}
	if err := p.parseVersionOrChannel(ref); err != nil {
		t.Fatal(err)
	}

	if ref.version == nil {
		t.Fatal("expected version, got nil")
	}
	if ref.channel != nil {
		t.Error("expected channel to be nil")
	}
}

func TestReferenceParser_ParseVersionOrChannel_Channel(t *testing.T) {
	p := &referenceParser{
		tokens: []string{":stable"},
		pos:    0,
	}

	ref := &Reference{}
	if err := p.parseVersionOrChannel(ref); err != nil {
		t.Fatal(err)
	}

	if ref.channel == nil {
		t.Fatal("expected channel, got nil")
	}
	if *ref.channel != "stable" {
		t.Errorf("expected channel %q, got %q", "stable", *ref.channel)
	}
	if ref.version != nil {
		t.Error("expected version to be nil")
	}
}

func TestReferenceParser_ParseVersionOrChannel_MultipleTokens(t *testing.T) {
	p := &referenceParser{
		tokens: []string{">=1.0.0", "<2.0.0"},
		pos:    0,
	}

	ref := &Reference{}
	if err := p.parseVersionOrChannel(ref); err != nil {
		t.Fatal(err)
	}

	if ref.version == nil {
		t.Fatal("expected version, got nil")
	}
}

func TestReferenceParser_ParseVersionOrChannel_StopsAtDigest(t *testing.T) {
	p := &referenceParser{
		tokens: []string{"1.0.0", "sha256:abcd1234"},
		pos:    0,
	}

	ref := &Reference{}
	if err := p.parseVersionOrChannel(ref); err != nil {
		t.Fatal(err)
	}

	if p.pos != 1 {
		t.Errorf("expected pos 1, got %d", p.pos)
	}
}

func TestReferenceParser_ParseVersionOrChannel_Empty(t *testing.T) {
	p := &referenceParser{
		tokens: []string{},
		pos:    0,
	}

	ref := &Reference{}
	err := p.parseVersionOrChannel(ref)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidReference) {
		t.Errorf("expected ErrInvalidReference, got %v", err)
	}
}

func TestReferenceParser_ParseDigest(t *testing.T) {
	p := &referenceParser{
		tokens: []string{"sha256:abcd1234"},
		pos:    0,
	}

	ref := &Reference{}
	if err := p.parseDigest(ref); err != nil {
		t.Fatal(err)
	}

	if ref.digest == nil {
		t.Fatal("expected digest, got nil")
	}
}

func TestReferenceParser_ParseDigest_Empty(t *testing.T) {
	p := &referenceParser{
		tokens: []string{},
		pos:    0,
	}

	ref := &Reference{}
	if err := p.parseDigest(ref); err != nil {
		t.Fatal(err)
	}

	if ref.digest != nil {
		t.Error("expected digest to be nil")
	}
}

func TestReferenceParser_ParseDigest_NotDigest(t *testing.T) {
	p := &referenceParser{
		tokens: []string{"notadigest"},
		pos:    0,
	}

	ref := &Reference{}
	if err := p.parseDigest(ref); err != nil {
		t.Fatal(err)
	}

	if ref.digest != nil {
		t.Error("expected digest to be nil")
	}
	if p.pos != 0 {
		t.Errorf("expected pos 0, got %d", p.pos)
	}
}

func TestLooksLikeVersion_GreaterThan(t *testing.T) {
	if !looksLikeVersion(">1.0.0") {
		t.Error("expected true for >1.0.0")
	}
}

func TestLooksLikeVersion_LessThan(t *testing.T) {
	if !looksLikeVersion("<2.0.0") {
		t.Error("expected true for <2.0.0")
	}
}

func TestLooksLikeVersion_Equals(t *testing.T) {
	if !looksLikeVersion("=1.0.0") {
		t.Error("expected true for =1.0.0")
	}
}

func TestLooksLikeVersion_Caret(t *testing.T) {
	if !looksLikeVersion("^1.0.0") {
		t.Error("expected true for ^1.0.0")
	}
}

func TestLooksLikeVersion_Tilde(t *testing.T) {
	if !looksLikeVersion("~1.0.0") {
		t.Error("expected true for ~1.0.0")
	}
}

func TestLooksLikeVersion_LowercaseV(t *testing.T) {
	if !looksLikeVersion("v1.0.0") {
		t.Error("expected true for v1.0.0")
	}
}

func TestLooksLikeVersion_UppercaseV(t *testing.T) {
	if !looksLikeVersion("V1.0.0") {
		t.Error("expected true for V1.0.0")
	}
}

func TestLooksLikeVersion_Digit(t *testing.T) {
	if !looksLikeVersion("1.0.0") {
		t.Error("expected true for 1.0.0")
	}
}

func TestLooksLikeVersion_Empty(t *testing.T) {
	if looksLikeVersion("") {
		t.Error("expected false for empty string")
	}
}

func TestLooksLikeVersion_Name(t *testing.T) {
	if looksLikeVersion("widget") {
		t.Error("expected false for widget")
	}
}

func TestLooksLikeVersion_Path(t *testing.T) {
	if looksLikeVersion("namespace/name") {
		t.Error("expected false for namespace/name")
	}
}
