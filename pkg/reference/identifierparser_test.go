package reference

import (
	"errors"
	"testing"
)

func TestIdentifierParser_Parse_NameOnly(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"widget"},
		options: nil,
	}

	id, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if id.typ != "template" {
		t.Errorf("expected type %q, got %q", "template", id.typ)
	}
	if id.scheme != DefaultScheme {
		t.Errorf("expected scheme %q, got %q", DefaultScheme, id.scheme)
	}
	if id.registry != DefaultRegistry {
		t.Errorf("expected registry %q, got %q", DefaultRegistry, id.registry)
	}
	if id.namespace != DefaultNamespace {
		t.Errorf("expected namespace %q, got %q", DefaultNamespace, id.namespace)
	}
	if id.name != "widget" {
		t.Errorf("expected name %q, got %q", "widget", id.name)
	}
}

func TestIdentifierParser_Parse_NamespaceAndName(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"namespace/name"},
		options: nil,
	}

	id, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if id.namespace != "namespace" {
		t.Errorf("expected namespace %q, got %q", "namespace", id.namespace)
	}
	if id.name != "name" {
		t.Errorf("expected name %q, got %q", "name", id.name)
	}
}

func TestIdentifierParser_Parse_WithExplicitType(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"template", "namespace/name"},
		options: nil,
	}

	id, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if id.typ != "template" {
		t.Errorf("expected type %q, got %q", "template", id.typ)
	}
	if id.namespace != "namespace" {
		t.Errorf("expected namespace %q, got %q", "namespace", id.namespace)
	}
}

func TestIdentifierParser_Parse_FullURI(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"https://myregistry.com/path/to/resource"},
		options: nil,
	}

	id, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if id.scheme != "https" {
		t.Errorf("expected scheme %q, got %q", "https", id.scheme)
	}
	if id.registry != "myregistry.com" {
		t.Errorf("expected registry %q, got %q", "myregistry.com", id.registry)
	}
	if id.path != "path/to/resource" {
		t.Errorf("expected path %q, got %q", "path/to/resource", id.path)
	}
}

func TestIdentifierParser_Parse_RegistryAndPath(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"myregistry.com/path/to/resource"},
		options: nil,
	}

	id, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if id.scheme != DefaultScheme {
		t.Errorf("expected scheme %q, got %q", DefaultScheme, id.scheme)
	}
	if id.registry != "myregistry.com" {
		t.Errorf("expected registry %q, got %q", "myregistry.com", id.registry)
	}
	if id.path != "path/to/resource" {
		t.Errorf("expected path %q, got %q", "path/to/resource", id.path)
	}
}

func TestIdentifierParser_Parse_WithOptions(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"widget"},
		options: &IdentifierOptions{
			DefaultScheme:    "oci",
			DefaultRegistry:  "custom.registry.io",
			DefaultNamespace: "myteam",
		},
	}

	id, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if id.scheme != "oci" {
		t.Errorf("expected scheme %q, got %q", "oci", id.scheme)
	}
	if id.registry != "custom.registry.io" {
		t.Errorf("expected registry %q, got %q", "custom.registry.io", id.registry)
	}
	if id.namespace != "myteam" {
		t.Errorf("expected namespace %q, got %q", "myteam", id.namespace)
	}
}

func TestIdentifierParser_Parse_InvalidContextType(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"widget"},
		options: nil,
	}

	_, err := p.parse("Invalid")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_Parse_EmptyTokens(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	_, err := p.parse("template")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_Parse_TypeMismatch(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"plugin", "namespace/name"},
		options: nil,
	}

	_, err := p.parse("template")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrTypeMismatch) {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

func TestIdentifierParser_Parse_UnexpectedToken(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"namespace/name", "extra"},
		options: nil,
	}

	_, err := p.parse("template")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_Peek(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"a", "b", "c"},
	}

	tok, ok := p.peek()
	if !ok || tok != "a" {
		t.Errorf("expected %q, got %q", "a", tok)
	}

	// Peek should not advance
	tok, ok = p.peek()
	if !ok || tok != "a" {
		t.Errorf("expected %q, got %q", "a", tok)
	}
}

func TestIdentifierParser_Peek_Empty(t *testing.T) {
	p := &identifierParser{
		tokens: []string{},
	}

	_, ok := p.peek()
	if ok {
		t.Error("expected ok to be false")
	}
}

func TestIdentifierParser_Next(t *testing.T) {
	p := &identifierParser{
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

func TestIdentifierParser_ParseType_ContextOnly(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"namespace/name"},
		options: nil,
	}

	id := &Identifier{}
	if err := p.parseType(id, "template"); err != nil {
		t.Fatal(err)
	}

	if id.typ != "template" {
		t.Errorf("expected type %q, got %q", "template", id.typ)
	}

	// Position should not advance
	if p.pos != 0 {
		t.Errorf("expected pos 0, got %d", p.pos)
	}
}

func TestIdentifierParser_ParseType_Explicit(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"template", "namespace/name"},
		options: nil,
	}

	id := &Identifier{}
	if err := p.parseType(id, "template"); err != nil {
		t.Fatal(err)
	}

	if id.typ != "template" {
		t.Errorf("expected type %q, got %q", "template", id.typ)
	}

	// Position should advance past type
	if p.pos != 1 {
		t.Errorf("expected pos 1, got %d", p.pos)
	}
}

func TestIdentifierParser_ParseType_Mismatch(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"plugin", "namespace/name"},
		options: nil,
	}

	id := &Identifier{}
	err := p.parseType(id, "template")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrTypeMismatch) {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

func TestIdentifierParser_ParseType_NonAlphabetic(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"123", "namespace/name"},
		options: nil,
	}

	id := &Identifier{}
	if err := p.parseType(id, "template"); err != nil {
		t.Fatal(err)
	}

	// Non-alphabetic token is not a type; position should not advance
	if p.pos != 0 {
		t.Errorf("expected pos 0, got %d", p.pos)
	}
}

func TestIdentifierParser_ParseType_SingleToken(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"widget"},
		options: nil,
	}

	id := &Identifier{}
	if err := p.parseType(id, "template"); err != nil {
		t.Fatal(err)
	}

	// Single token is a path, not a type
	if p.pos != 0 {
		t.Errorf("expected pos 0, got %d", p.pos)
	}
}

func TestIdentifierParser_ParseLocation_DefaultPath(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"widget"},
		options: nil,
	}

	id := &Identifier{}
	if err := p.parseLocation(id); err != nil {
		t.Fatal(err)
	}

	if id.namespace != DefaultNamespace {
		t.Errorf("expected namespace %q, got %q", DefaultNamespace, id.namespace)
	}
	if id.name != "widget" {
		t.Errorf("expected name %q, got %q", "widget", id.name)
	}
}

func TestIdentifierParser_ParseLocation_NamespaceAndName(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{"namespace/name"},
		options: nil,
	}

	id := &Identifier{}
	if err := p.parseLocation(id); err != nil {
		t.Fatal(err)
	}

	if id.namespace != "namespace" {
		t.Errorf("expected namespace %q, got %q", "namespace", id.namespace)
	}
	if id.name != "name" {
		t.Errorf("expected name %q, got %q", "name", id.name)
	}
}

func TestIdentifierParser_ParseLocation_Empty(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	err := p.parseLocation(id)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_ParseURI(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	if err := p.parseURI(id, "https", "myregistry.com/path/to/resource"); err != nil {
		t.Fatal(err)
	}

	if id.scheme != "https" {
		t.Errorf("expected scheme %q, got %q", "https", id.scheme)
	}
	if id.registry != "myregistry.com" {
		t.Errorf("expected registry %q, got %q", "myregistry.com", id.registry)
	}
	if id.path != "path/to/resource" {
		t.Errorf("expected path %q, got %q", "path/to/resource", id.path)
	}
}

func TestIdentifierParser_ParseURI_InvalidScheme(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	err := p.parseURI(id, "123", "myregistry.com/path")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_ParseURI_InvalidRegistry(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	err := p.parseURI(id, "https", "-invalid/path")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_ParseURI_InvalidPath(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	err := p.parseURI(id, "https", "myregistry.com/INVALID")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_ParseURI_MissingPath(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	err := p.parseURI(id, "https", "myregistry.com")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_ParseURI_EmptyPath(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	err := p.parseURI(id, "https", "myregistry.com/")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_ParseRegistryPath(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	if err := p.parseRegistryPath(id, "myregistry.com", "path/to/resource"); err != nil {
		t.Fatal(err)
	}

	if id.scheme != DefaultScheme {
		t.Errorf("expected scheme %q, got %q", DefaultScheme, id.scheme)
	}
	if id.registry != "myregistry.com" {
		t.Errorf("expected registry %q, got %q", "myregistry.com", id.registry)
	}
	if id.path != "path/to/resource" {
		t.Errorf("expected path %q, got %q", "path/to/resource", id.path)
	}
}

func TestIdentifierParser_ParseRegistryPath_WithOptions(t *testing.T) {
	p := &identifierParser{
		tokens: []string{},
		options: &IdentifierOptions{
			DefaultScheme: "oci",
		},
	}

	id := &Identifier{}
	if err := p.parseRegistryPath(id, "myregistry.com", "path/to/resource"); err != nil {
		t.Fatal(err)
	}

	if id.scheme != "oci" {
		t.Errorf("expected scheme %q, got %q", "oci", id.scheme)
	}
}

func TestIdentifierParser_ParseRegistryPath_InvalidRegistry(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	err := p.parseRegistryPath(id, "-invalid", "path")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_ParseRegistryPath_EmptyPath(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	err := p.parseRegistryPath(id, "myregistry.com", "")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_ParseRegistryPath_InvalidPath(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	err := p.parseRegistryPath(id, "myregistry.com", "INVALID")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_ParseDefaultPath_NameOnly(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	if err := p.parseDefaultPath(id, "widget"); err != nil {
		t.Fatal(err)
	}

	if id.namespace != DefaultNamespace {
		t.Errorf("expected namespace %q, got %q", DefaultNamespace, id.namespace)
	}
	if id.name != "widget" {
		t.Errorf("expected name %q, got %q", "widget", id.name)
	}
}

func TestIdentifierParser_ParseDefaultPath_WithNamespace(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	if err := p.parseDefaultPath(id, "namespace/name"); err != nil {
		t.Fatal(err)
	}

	if id.namespace != "namespace" {
		t.Errorf("expected namespace %q, got %q", "namespace", id.namespace)
	}
	if id.name != "name" {
		t.Errorf("expected name %q, got %q", "name", id.name)
	}
}

func TestIdentifierParser_ParseDefaultPath_WithOptions(t *testing.T) {
	p := &identifierParser{
		tokens: []string{},
		options: &IdentifierOptions{
			DefaultScheme:    "oci",
			DefaultRegistry:  "custom.registry.io",
			DefaultNamespace: "namespace",
		},
	}

	id := &Identifier{}
	if err := p.parseDefaultPath(id, "name"); err != nil {
		t.Fatal(err)
	}

	if id.scheme != "oci" {
		t.Errorf("expected scheme %q, got %q", "oci", id.scheme)
	}
	if id.registry != "custom.registry.io" {
		t.Errorf("expected registry %q, got %q", "custom.registry.io", id.registry)
	}
	if id.namespace != "namespace" {
		t.Errorf("expected namespace %q, got %q", "namespace", id.namespace)
	}
}

func TestIdentifierParser_ParseDefaultPath_InvalidName(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	err := p.parseDefaultPath(id, "Invalid")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_ParseDefaultPath_InvalidNamespace(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	err := p.parseDefaultPath(id, "Invalid/widget")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_ParseDefaultPath_InvalidNameInNamespace(t *testing.T) {
	p := &identifierParser{
		tokens:  []string{},
		options: nil,
	}

	id := &Identifier{}
	err := p.parseDefaultPath(id, "myteam/Invalid")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestLooksLikeRegistry_WithDot(t *testing.T) {
	if !looksLikeRegistry("myregistry.com") {
		t.Error("expected true for myregistry.com")
	}
}

func TestLooksLikeRegistry_WithPort(t *testing.T) {
	if !looksLikeRegistry("localhost:8080") {
		t.Error("expected true for localhost:8080")
	}
}

func TestLooksLikeRegistry_WithDotAndPort(t *testing.T) {
	if !looksLikeRegistry("myregistry.com:443") {
		t.Error("expected true for myregistry.com:443")
	}
}

func TestLooksLikeRegistry_Subdomain(t *testing.T) {
	if !looksLikeRegistry("sub.myregistry.com") {
		t.Error("expected true for sub.myregistry.com")
	}
}

func TestLooksLikeRegistry_SimpleName(t *testing.T) {
	if looksLikeRegistry("myteam") {
		t.Error("expected false for myteam")
	}
}

func TestLooksLikeRegistry_Widget(t *testing.T) {
	if looksLikeRegistry("widget") {
		t.Error("expected false for widget")
	}
}
