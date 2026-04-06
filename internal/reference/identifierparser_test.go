package reference

import (
	"errors"
	"testing"
)

func TestIdentifierParser_Parse_NameOnly(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"widget"},
	}

	id, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if id.typ != "template" {
		t.Errorf("expected type %q, got %q", "template", id.typ)
	}
	if id.namespace != "" {
		t.Errorf("expected empty namespace, got %q", id.namespace)
	}
	if id.name != "widget" {
		t.Errorf("expected name %q, got %q", "widget", id.name)
	}
}

func TestIdentifierParser_Parse_NamespaceAndName(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"namespace/name"},
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

func TestIdentifierParser_Parse_RegistryNamespaceAndName(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"hub.example.com/namespace/name"},
	}

	id, err := p.parse("template")
	if err != nil {
		t.Fatal(err)
	}

	if id.registry != "hub.example.com" {
		t.Errorf("expected registry %q, got %q", "hub.example.com", id.registry)
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
		tokens: []string{"template", "namespace/name"},
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

func TestIdentifierParser_Parse_InvalidContextType(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"widget"},
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
		tokens: []string{},
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
		tokens: []string{"plugin", "namespace/name"},
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
		tokens: []string{"namespace/name", "extra"},
	}

	_, err := p.parse("template")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("expected ErrInvalidIdentifier, got %v", err)
	}
}

func TestIdentifierParser_Parse_TooManySegments(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"a/b/c/d"},
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
		tokens: []string{"namespace/name"},
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
		tokens: []string{"template", "namespace/name"},
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
		tokens: []string{"plugin", "namespace/name"},
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
		tokens: []string{"123", "namespace/name"},
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
		tokens: []string{"widget"},
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

func TestIdentifierParser_ParseLocation_NameOnly(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"widget"},
	}

	id := &Identifier{}
	if err := p.parseLocation(id); err != nil {
		t.Fatal(err)
	}

	if id.registry != "" {
		t.Errorf("expected empty registry, got %q", id.registry)
	}
	if id.namespace != "" {
		t.Errorf("expected empty namespace, got %q", id.namespace)
	}
	if id.name != "widget" {
		t.Errorf("expected name %q, got %q", "widget", id.name)
	}
}

func TestIdentifierParser_ParseLocation_NamespaceAndName(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"namespace/name"},
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

func TestIdentifierParser_ParseLocation_RegistryNamespaceAndName(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"hub.example.com/namespace/name"},
	}

	id := &Identifier{}
	if err := p.parseLocation(id); err != nil {
		t.Fatal(err)
	}

	if id.registry != "hub.example.com" {
		t.Errorf("expected registry %q, got %q", "hub.example.com", id.registry)
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
		tokens: []string{},
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

func TestIdentifierParser_ParseLocation_InvalidName(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"Invalid"},
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

func TestIdentifierParser_ParseLocation_InvalidNamespace(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"Invalid/widget"},
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

func TestIdentifierParser_ParseLocation_InvalidNameInNamespace(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"myteam/Invalid"},
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

func TestIdentifierParser_ParseLocation_TooManySegments(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"a/b/c/d"},
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

func TestIdentifierParser_ParseLocation_EmptyRegistry(t *testing.T) {
	p := &identifierParser{
		tokens: []string{"/namespace/name"},
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
