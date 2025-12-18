package reference

import (
	"fmt"
	"regexp"
	"strings"
)

var (

	// Channel: colon prefix, lowercase alphanumeric with hyphens.
	channelPattern = regexp.MustCompile(`^:[a-z][a-z0-9-]*$`)

	// Digest: algorithm:hexhash (e.g., sha256:abc123).
	digestPattern = regexp.MustCompile(`^[a-z0-9]+:[a-f0-9]+$`)
)

// Whitespace-tokenized reference string parser.
type referenceParser struct {
	tokens  []string
	pos     int
	options *IdentifierOptions
}

// Parses the tokens into a Reference.
func (p *referenceParser) parse(contextType string) (*Reference, error) {

	// Find where the identifier ends and version/channel begins
	idEnd := p.findIdentifierEnd()
	if idEnd == 0 {
		return nil, fmt.Errorf("%w: empty reference", ErrInvalidReference)
	}

	// Parse the identifier portion
	idParser := &identifierParser{
		tokens:  p.tokens[:idEnd],
		options: p.options,
	}

	id, err := idParser.parse(contextType)
	if err != nil {
		return nil, err
	}

	ref := &Reference{
		Identifier: *id,
	}

	// Advance past identifier tokens
	p.pos = idEnd

	if err := p.parseVersionOrChannel(ref); err != nil {
		return nil, err
	}

	if err := p.parseDigest(ref); err != nil {
		return nil, err
	}

	if tok, ok := p.peek(); ok {
		return nil, fmt.Errorf("%w: unexpected token %q", ErrInvalidReference, tok)
	}

	return ref, nil
}

// Finds the end position of identifier tokens.
//
// The identifier ends when we encounter a version constraint, channel, or digest.
func (p *referenceParser) findIdentifierEnd() int {
	for i, tok := range p.tokens {
		if channelPattern.MatchString(tok) {
			return i
		}
		if digestPattern.MatchString(tok) {
			return i
		}
		if looksLikeVersion(tok) {
			return i
		}
	}
	return len(p.tokens)
}

// Returns the current token without advancing.
func (p *referenceParser) peek() (string, bool) {
	if p.pos >= len(p.tokens) {
		return "", false
	}
	return p.tokens[p.pos], true
}

// Returns the current token and advances.
func (p *referenceParser) next() (string, bool) {
	tok, ok := p.peek()
	if ok {
		p.pos++
	}
	return tok, ok
}

// Returns the number of remaining tokens.
func (p *referenceParser) remaining() int {
	return len(p.tokens) - p.pos
}

// Parses version constraint or channel (required).
func (p *referenceParser) parseVersionOrChannel(ref *Reference) error {
	tok, ok := p.peek()
	if !ok {
		return fmt.Errorf("%w: missing version or channel", ErrInvalidReference)
	}

	// Check for channel
	if channelPattern.MatchString(tok) {
		channel := strings.TrimPrefix(tok, ":")
		ref.channel = &channel
		p.next()
		return nil
	}

	// Parse version constraint (may span multiple tokens)
	var versionTokens []string
	for p.remaining() > 0 {
		tok, _ := p.peek()
		if digestPattern.MatchString(tok) {
			break
		}
		versionTokens = append(versionTokens, tok)
		p.next()
	}

	if len(versionTokens) == 0 {
		return fmt.Errorf("%w: missing version or channel", ErrInvalidReference)
	}

	versionStr := strings.Join(versionTokens, " ")
	constraint, err := ParseVersionConstraint(versionStr)
	if err != nil {
		return err
	}

	ref.version = constraint
	return nil
}

// Parses the optional digest suffix.
func (p *referenceParser) parseDigest(ref *Reference) error {
	tok, ok := p.peek()
	if !ok {
		return nil
	}

	if !digestPattern.MatchString(tok) {
		return nil
	}

	p.next()
	digest, err := ParseDigest(tok)
	if err != nil {
		return err
	}

	ref.digest = digest
	return nil
}

// Returns true if the string looks like a version constraint.
func looksLikeVersion(s string) bool {
	if len(s) == 0 {
		return false
	}
	c := s[0]
	return c == '>' || c == '<' || c == '=' || c == '^' || c == '~' ||
		c == 'v' || c == 'V' || (c >= '0' && c <= '9')
}
