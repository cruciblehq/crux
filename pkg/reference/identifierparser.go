package reference

import (
	"fmt"
	"regexp"
	"strings"
)

var (

	// Type: lowercase alphabetic only.
	typePattern = regexp.MustCompile(`^[a-z]+$`)

	// Scheme: lowercase alphabetic followed by optional digits, plus, dot, or hyphen.
	schemePattern = regexp.MustCompile(`^[a-z][a-z0-9+.-]*$`)

	// Registry: alphanumeric, starting/ending with alphanumeric, separated by
	// dots or hyphens. May end with colon and port.
	registryPattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?))+\.?(:\d+)?$`)

	// Name: lowercase alphanumeric with hyphens, starting with letter.
	namePattern = regexp.MustCompile(`^[a-z]([a-z0-9-]{0,126}[a-z0-9])?$`)

	// Path: lowercase, digits, hyphens, slashes, underscores, dots.
	pathPattern = regexp.MustCompile(`^[a-z0-9/_.-]+$`)
)

// Whitespace-tokenized identifier string parser.
type identifierParser struct {
	tokens  []string           // Tokenized input
	pos     int                // Parser position in tokens
	options *IdentifierOptions // Parsing options
}

// Parses the tokens into an Identifier.
func (p *identifierParser) parse(contextType string) (*Identifier, error) {
	if !typePattern.MatchString(contextType) {
		return nil, fmt.Errorf("%w: invalid type context %q", ErrInvalidIdentifier, contextType)
	}

	if len(p.tokens) == 0 {
		return nil, fmt.Errorf("%w: empty identifier", ErrInvalidIdentifier)
	}

	id := &Identifier{
		scheme:   p.options.scheme(),
		registry: p.options.registry(),
	}

	if err := p.parseType(id, contextType); err != nil {
		return nil, err
	}

	if err := p.parseLocation(id); err != nil {
		return nil, err
	}

	if tok, ok := p.peek(); ok {
		return nil, fmt.Errorf("%w: unexpected token %q", ErrInvalidIdentifier, tok)
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

	// Look ahead: if next looks like a path, current is type not path
	if p.pos+1 < len(p.tokens) {
		next := p.tokens[p.pos+1]
		if !strings.Contains(next, "/") && !looksLikeRegistry(next) {
			return nil
		}
	} else {
		// Single token remaining; it's a path, not a type
		return nil
	}

	// Token is a type; must match context.
	if tok != contextType {
		return fmt.Errorf("%w: type %q does not match context %q", ErrTypeMismatch, tok, contextType)
	}
	p.pos++

	return nil
}

// Parses the resource location (scheme, registry, path).
func (p *identifierParser) parseLocation(id *Identifier) error {
	tok, ok := p.next()
	if !ok {
		return fmt.Errorf("%w: missing resource location", ErrInvalidIdentifier)
	}

	// Full URI: scheme://registry/path
	if scheme, rest, ok := strings.Cut(tok, "://"); ok {
		return p.parseURI(id, scheme, rest)
	}

	// Check if first segment looks like a registry
	if first, rest, ok := strings.Cut(tok, "/"); ok && looksLikeRegistry(first) {
		return p.parseRegistryPath(id, first, rest)
	}

	// Default registry: namespace/name or name
	return p.parseDefaultPath(id, tok)
}

// Parses a full URI (scheme://registry/path).
func (p *identifierParser) parseURI(id *Identifier, scheme, rest string) error {
	if !schemePattern.MatchString(scheme) {
		return fmt.Errorf("%w: invalid scheme %q", ErrInvalidIdentifier, scheme)
	}

	registry, path, ok := strings.Cut(rest, "/")
	if !ok || registry == "" || path == "" {
		return fmt.Errorf("%w: invalid URI", ErrInvalidIdentifier)
	}

	if !registryPattern.MatchString(registry) {
		return fmt.Errorf("%w: invalid registry %q", ErrInvalidIdentifier, registry)
	}

	if !pathPattern.MatchString(path) {
		return fmt.Errorf("%w: invalid path %q", ErrInvalidIdentifier, path)
	}

	id.scheme = scheme
	id.registry = registry
	id.path = path

	return nil
}

// Parses a registry/path combination without scheme.
func (p *identifierParser) parseRegistryPath(id *Identifier, registry, path string) error {
	if !registryPattern.MatchString(registry) {
		return fmt.Errorf("%w: invalid registry %q", ErrInvalidIdentifier, registry)
	}

	if path == "" {
		return fmt.Errorf("%w: empty path", ErrInvalidIdentifier)
	}

	if !pathPattern.MatchString(path) {
		return fmt.Errorf("%w: invalid path %q", ErrInvalidIdentifier, path)
	}

	id.scheme = p.options.scheme()
	id.registry = registry
	id.path = path

	return nil
}

// Parses a default registry path (namespace/name or just name).
func (p *identifierParser) parseDefaultPath(id *Identifier, tok string) error {
	id.scheme = p.options.scheme()
	id.registry = p.options.registry()

	if namespace, name, ok := strings.Cut(tok, "/"); ok {
		if !namePattern.MatchString(namespace) {
			return fmt.Errorf("%w: invalid namespace %q", ErrInvalidIdentifier, namespace)
		}
		if !namePattern.MatchString(name) {
			return fmt.Errorf("%w: invalid name %q", ErrInvalidIdentifier, name)
		}
		id.namespace = namespace
		id.name = name
	} else {
		if !namePattern.MatchString(tok) {
			return fmt.Errorf("%w: invalid name %q", ErrInvalidIdentifier, tok)
		}
		id.namespace = p.options.namespace()
		id.name = tok
	}

	return nil
}

// Returns true if the string looks like a registry hostname.
func looksLikeRegistry(s string) bool {
	return strings.Contains(s, ".") || strings.Contains(s, ":")
}
