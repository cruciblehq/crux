package reference

import (
	"fmt"
	"regexp"
	"strings"
)

var (

	// Type: lowercase alphabetic only.
	typePattern = regexp.MustCompile(`^[a-z]+$`)

	// Name: lowercase alphanumeric with hyphens, starting with letter.
	namePattern = regexp.MustCompile(`^[a-z]([a-z0-9-]{0,126}[a-z0-9])?$`)
)

// Whitespace-tokenized identifier string parser.
type identifierParser struct {
	tokens []string // Tokenized input
	pos    int      // Parser position in tokens
}

// Parses the tokens into an Identifier.
func (p *identifierParser) parse(contextType string) (*Identifier, error) {
	if !typePattern.MatchString(contextType) {
		return nil, wrap(ErrInvalidIdentifier, ErrInvalidContextType)
	}

	if len(p.tokens) == 0 {
		return nil, wrap(ErrInvalidIdentifier, ErrEmptyIdentifier)
	}

	id := &Identifier{}

	if err := p.parseType(id, contextType); err != nil {
		return nil, err
	}

	if err := p.parseLocation(id); err != nil {
		return nil, err
	}

	if _, ok := p.peek(); ok {
		return nil, wrap(ErrInvalidIdentifier, ErrUnexpectedToken)
	}

	return id, nil
}

// Returns the current token without advancing.
func (p *identifierParser) peek() (string, bool) {
	if p.pos >= len(p.tokens) {
		return "", false
	}
	return p.tokens[p.pos], true
}

// Returns the current token and advances.
func (p *identifierParser) next() (string, bool) {
	tok, ok := p.peek()
	if ok {
		p.pos++
	}
	return tok, ok
}

// Parses the optional type prefix.
func (p *identifierParser) parseType(id *Identifier, contextType string) error {
	id.typ = contextType

	tok, ok := p.peek()
	if !ok || !typePattern.MatchString(tok) {
		return nil
	}

	// Look ahead: if next token contains a slash, current is type not path.
	if p.pos+1 < len(p.tokens) {
		next := p.tokens[p.pos+1]
		if !strings.Contains(next, "/") {
			return nil
		}
	} else {
		// Single token remaining; it's a path, not a type.
		return nil
	}

	// Token is a type; must match context.
	if tok != contextType {
		return wrap(ErrTypeMismatch, fmt.Errorf("type %q does not match context %q", tok, contextType))
	}
	p.pos++

	return nil
}

// Parses the resource location.
//
// More than 3 segments is an error. Registry and namespace may be empty in
// the parsed result — callers are expected to apply defaults when needed.
func (p *identifierParser) parseLocation(id *Identifier) error {
	tok, ok := p.next()
	if !ok {
		return wrap(ErrInvalidIdentifier, ErrEmptyIdentifier)
	}

	parts := strings.Split(tok, "/")

	switch len(parts) {
	case 1:
		// name
		if !namePattern.MatchString(parts[0]) {
			return wrap(ErrInvalidIdentifier, ErrInvalidName)
		}
		id.name = parts[0]

	case 2:
		// namespace/name
		if !namePattern.MatchString(parts[0]) {
			return wrap(ErrInvalidIdentifier, ErrInvalidNamespace)
		}
		if !namePattern.MatchString(parts[1]) {
			return wrap(ErrInvalidIdentifier, ErrInvalidName)
		}
		id.namespace = parts[0]
		id.name = parts[1]

	case 3:
		// registry/namespace/name
		if parts[0] == "" {
			return wrap(ErrInvalidIdentifier, ErrInvalidRegistry)
		}
		if !namePattern.MatchString(parts[1]) {
			return wrap(ErrInvalidIdentifier, ErrInvalidNamespace)
		}
		if !namePattern.MatchString(parts[2]) {
			return wrap(ErrInvalidIdentifier, ErrInvalidName)
		}
		id.registry = parts[0]
		id.namespace = parts[1]
		id.name = parts[2]

	default:
		return wrap(ErrInvalidIdentifier, ErrInvalidPath)
	}

	return nil
}
