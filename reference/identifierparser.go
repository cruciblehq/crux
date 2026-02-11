package reference

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/cruciblehq/crux/kit/crex"
	"github.com/cruciblehq/crux/resource"
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
	tokens  []string          // Tokenized input
	pos     int               // Parser position in tokens
	options IdentifierOptions // Parsing options
}

// Parses the tokens into an Identifier.
func (p *identifierParser) parse(contextType resource.Type) (*Identifier, error) {
	if !typePattern.MatchString(string(contextType)) {
		return nil, crex.Wrap(ErrInvalidIdentifier, ErrInvalidContextType)
	}

	if len(p.tokens) == 0 {
		return nil, crex.Wrap(ErrInvalidIdentifier, ErrEmptyIdentifier)
	}

	id := &Identifier{}

	if err := p.parseType(id, contextType); err != nil {
		return nil, err
	}

	if err := p.parseLocation(id); err != nil {
		return nil, err
	}

	if _, ok := p.peek(); ok {
		return nil, crex.Wrap(ErrInvalidIdentifier, ErrUnexpectedToken)
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
func (p *identifierParser) parseType(id *Identifier, contextType resource.Type) error {
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
	if tok != string(contextType) {
		return crex.Wrap(ErrTypeMismatch, fmt.Errorf("type %q does not match context %q", tok, contextType))
	}
	p.pos++

	return nil
}

// Parses the resource location (scheme, registry, path).
func (p *identifierParser) parseLocation(id *Identifier) error {
	tok, ok := p.next()
	if !ok {
		return crex.Wrap(ErrInvalidIdentifier, ErrEmptyIdentifier)
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
		return crex.Wrap(ErrInvalidIdentifier, ErrInvalidScheme)
	}

	registry, path, ok := strings.Cut(rest, "/")
	if !ok || registry == "" || path == "" {
		if !ok || registry == "" {
			return crex.Wrap(ErrInvalidIdentifier, ErrMissingRegistry)
		}
		return crex.Wrap(ErrInvalidIdentifier, ErrMissingPath)
	}

	if !registryPattern.MatchString(registry) {
		return crex.Wrap(ErrInvalidIdentifier, ErrInvalidRegistry)
	}

	if !pathPattern.MatchString(path) {
		return crex.Wrap(ErrInvalidIdentifier, ErrInvalidPath)
	}

	id.registry = &url.URL{
		Scheme: scheme,
		Host:   registry,
	}
	id.path = path

	return nil
}

// Parses a registry/path combination without scheme.
func (p *identifierParser) parseRegistryPath(id *Identifier, registry, path string) error {
	if !registryPattern.MatchString(registry) {
		return crex.Wrap(ErrInvalidIdentifier, ErrInvalidRegistry)
	}

	if path == "" {
		return crex.Wrap(ErrInvalidIdentifier, ErrEmptyPath)
	}

	if !pathPattern.MatchString(path) {
		return crex.Wrap(ErrInvalidIdentifier, ErrInvalidPath)
	}

	id.registry = &url.URL{
		Scheme: "https",
		Host:   registry,
	}
	id.path = path

	return nil
}

// Parses a default registry path (namespace/name or just name).
func (p *identifierParser) parseDefaultPath(id *Identifier, tok string) error {
	u, err := url.Parse(p.options.DefaultRegistry)
	if err != nil {
		return crex.Wrap(ErrInvalidIdentifier, err)
	}
	id.registry = u

	if namespace, name, ok := strings.Cut(tok, "/"); ok {
		if !namePattern.MatchString(namespace) {
			return crex.Wrap(ErrInvalidIdentifier, ErrInvalidNamespace)
		}
		if !namePattern.MatchString(name) {
			return crex.Wrap(ErrInvalidIdentifier, ErrInvalidName)
		}
		id.namespace = namespace
		id.name = name
	} else {
		if !namePattern.MatchString(tok) {
			return crex.Wrap(ErrInvalidIdentifier, ErrInvalidName)
		}
		id.namespace = p.options.DefaultNamespace
		id.name = tok
	}

	return nil
}

// Returns true if the string looks like a registry hostname.
func looksLikeRegistry(s string) bool {
	return strings.Contains(s, ".") || strings.Contains(s, ":")
}
