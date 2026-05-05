package aegis

import (
	"strings"

	"github.com/cruciblehq/crux/internal/crex"
)

// Classifies a token produced by the lexer.
//
// Tokens are classified into types according to delimiting characters, where
// they appear in the source, and reserved keywords. Each type corresponds to
// a distinct lexical unit in the source text.
type TokenType string

const (
	TokenSubsystem TokenType = "subsystem" // Subsystem header (.name).
	TokenName      TokenType = "name"      // Name (identifier).
	TokenString    TokenType = "string"    // String literal (quoted or unquoted /path form).
	TokenInt       TokenType = "integer"   // Unsigned integer literal.
	TokenQuantity  TokenType = "quantity"  // Unsigned integer literal with a unit suffix (e.g. 1Gi, 500m).
	TokenVar       TokenType = "var"       // Variable reference.
	TokenWhere     TokenType = "where"     // Keyword separating rule header from expression.
	TokenAnd       TokenType = "and"       // Boolean and.
	TokenOr        TokenType = "or"        // Boolean or.
	TokenNot       TokenType = "not"       // Boolean not or negation.
	TokenIn        TokenType = "in"        // Set membership operator.
	TokenLike      TokenType = "like"      // Pattern matching operator.
	TokenBetween   TokenType = "between"   // Range operator.
	TokenEq        TokenType = "="         // Equality operator.
	TokenNeq       TokenType = "!="        // Inequality operator.
	TokenGt        TokenType = ">"         // Greater than operator.
	TokenGte       TokenType = ">="        // Greater than or equal operator.
	TokenLt        TokenType = "<"         // Less than operator.
	TokenLte       TokenType = "<="        // Less than or equal operator.
	TokenAmpersand TokenType = "&"         // Bitwise AND operator.
	TokenMinus     TokenType = "-"         // Minus or range dash.
	TokenLParen    TokenType = "("         // Left parenthesis.
	TokenRParen    TokenType = ")"         // Right parenthesis.
	TokenComma     TokenType = ","         // Comma separator.
	TokenError     TokenType = "error"     // Error token.
	TokenEOF       TokenType = "EOF"       // End of file.
)

// Declares which character set a token's value was written in.
//
// The lexer distinguishes between ASCII and Unicode for string literals,
// producing Unicode tokens only when the string is prefixed with a u""
// expression marker. For all other token types, the encoding is EncodingNone
// since the concept does not apply.
type TokenEncoding string

const (
	EncodingNone    TokenEncoding = "none"    // No encoding (not a string token).
	EncodingASCII   TokenEncoding = "ascii"   // ASCII character set.
	EncodingUnicode TokenEncoding = "unicode" // Unicode character set.
)

// A single lexical token.
//
// A token is a contiguous span of source text classified by Type. Text holds
// the verbatim source bytes the token covers, including surrounding quotes
// and unresolved escape sequences for string literals and the leading 'u' for
// Unicode string literals. Use Unquote to obtain the decoded content of a
// string token. For all other token types Text is already the value to use.
// Encoding is EncodingASCII or EncodingUnicode for TokenString and
// EncodingNone for all other token types.
type Token struct {
	Type     TokenType     // Token classification.
	Encoding TokenEncoding // Character set of the source text.
	Text     string        // Verbatim source slice covered by the token.
}

// Returns the decoded value of t.
//
// For TokenString tokens the surrounding quotes (and the leading 'u' for
// Unicode literals) are stripped and the recognised escape sequences (\"
// and \\) are resolved. For unquoted glob string tokens the source slice
// is returned unchanged. For all other token types Text is returned unchanged.
// Returns ErrLex wrapped with detail when the token is malformed.
func (t Token) Unquote() (string, error) {
	if t.Type != TokenString {
		return t.Text, nil
	}
	if len(t.Text) == 0 {
		return "", crex.Wrapf(ErrLex, "empty string token")
	}
	if t.Text[0] == '/' {
		return t.Text, nil // Unquoted glob; no quotes, no escapes.
	}
	body, err := stripQuotes(t.Text)
	if err != nil {
		return "", err
	}
	return unescapeBody(t.Text, body)
}

// Strips quotes and resolves escapes in a string token.
//
// Removes the optional 'u' prefix and surrounding double quotes from src and
// returns the inner body. Returns ErrLex wrapped with detail if src is not a
// well-formed quoted string. Used only by Unquote on TokenString values whose
// Text does not begin with '/'.
func stripQuotes(src string) (string, error) {
	s := src
	if s[0] == 'u' {
		s = s[1:]
	}
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return "", crex.Wrapf(ErrLex, "malformed quoted string %q", src)
	}
	return s[1 : len(s)-1], nil
}

// Resolves escape sequences in body and returns the decoded text.
//
// Recognised escape sequences are \" and \\. Returns ErrLex wrapped with
// detail for an unterminated or unknown escape. src is the original token
// text, used only for error messages. Used only by Unquote on TokenString
// values whose Text does not begin with '/'.
func unescapeBody(src, body string) (string, error) {
	var buf strings.Builder
	buf.Grow(len(body))
	for i := 0; i < len(body); i++ {
		ch := body[i]
		if ch != '\\' {
			buf.WriteByte(ch)
			continue
		}
		i++
		if i >= len(body) {
			return "", crex.Wrapf(ErrLex, "unterminated escape in %q", src)
		}
		esc := body[i]
		switch esc {
		case '"', '\\':
			buf.WriteByte(esc)
		default:
			return "", crex.Wrapf(ErrLex, "unknown escape \\%c in %q", esc, src)
		}
	}
	return buf.String(), nil
}

// Unit suffix on a quantity literal.
//
// The lexer recognises a closed set of suffixes that follow a decimal integer
// to form a single TokenQuantity (e.g. "1Gi", "500m"). The underlying string
// is the canonical source spelling, so a QuantitySuffix can be formatted
// directly with %s and round-trips through ParseQuantity. Each subsystem
// decides which suffixes a given knob accepts and what each suffix multiplies
// by (bytes, millicores, seconds, ...); the lexer attaches no semantics.
type QuantitySuffix string

// IEC binary multipliers. Conventionally power-of-1024 byte counts.
const (
	SuffixKi QuantitySuffix = "Ki" // 2^10
	SuffixMi QuantitySuffix = "Mi" // 2^20
	SuffixGi QuantitySuffix = "Gi" // 2^30
	SuffixTi QuantitySuffix = "Ti" // 2^40
	SuffixPi QuantitySuffix = "Pi" // 2^50
	SuffixEi QuantitySuffix = "Ei" // 2^60
)

// SI decimal multipliers. Conventionally power-of-1000 counts. Both the
// lower-case and upper-case forms of "kilo" are accepted.
const (
	SuffixKLower QuantitySuffix = "k" // 10^3
	SuffixK      QuantitySuffix = "K" // 10^3
	SuffixM      QuantitySuffix = "M" // 10^6
	SuffixG      QuantitySuffix = "G" // 10^9
	SuffixT      QuantitySuffix = "T" // 10^12
	SuffixP      QuantitySuffix = "P" // 10^15
	SuffixE      QuantitySuffix = "E" // 10^18
)

// SI sub-unit multipliers. Conventionally fractions of the base unit.
const (
	SuffixMilli QuantitySuffix = "m" // 10^-3
	SuffixMicro QuantitySuffix = "u" // 10^-6
	SuffixNano  QuantitySuffix = "n" // 10^-9
)

// Set of suffixes the lexer accepts. The lexer admits any of these after a
// decimal integer; subsystems decide which knobs accept which suffixes.
var quantitySuffixes = map[QuantitySuffix]struct{}{
	SuffixKi: {}, SuffixMi: {}, SuffixGi: {}, SuffixTi: {}, SuffixPi: {}, SuffixEi: {},
	SuffixKLower: {}, SuffixK: {}, SuffixM: {}, SuffixG: {}, SuffixT: {}, SuffixP: {}, SuffixE: {},
	SuffixMilli: {}, SuffixMicro: {}, SuffixNano: {},
}

// Whether s is a recognised quantity suffix.
func IsQuantitySuffix(s string) bool {
	_, ok := quantitySuffixes[QuantitySuffix(s)]
	return ok
}

// Splits a quantity literal into its numeric value and unit suffix.
//
// s must be a string of decimal digits followed by one of the recognised
// QuantitySuffix values (e.g. "1Gi", "500m", "100u"). Returns the parsed
// integer value, the suffix, and a nil error on success. Returns ErrLex
// wrapped with detail if s is malformed or carries an unknown suffix.
func ParseQuantity(s string) (uint64, QuantitySuffix, error) {
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, "", crex.Wrapf(ErrLex, "quantity %q has no digits", s)
	}
	suffix := QuantitySuffix(s[end:])
	if _, ok := quantitySuffixes[suffix]; !ok {
		return 0, "", crex.Wrapf(ErrLex, "quantity %q has unknown suffix %q", s, string(suffix))
	}
	v, err := parseDecimalUint(s[:end])
	if err != nil {
		return 0, "", crex.Wrapf(ErrLex, "quantity %q has invalid digits: %w", s, err)
	}
	return v, suffix, nil
}

// Parses a base-10 unsigned integer.
//
// Used by ParseQuantity to decode the digit prefix of a quantity literal.
// Returns ErrLex wrapped with detail if s is empty or contains a non-digit.
func parseDecimalUint(s string) (uint64, error) {
	if len(s) == 0 {
		return 0, crex.Wrapf(ErrLex, "empty integer")
	}
	var v uint64
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch < '0' || ch > '9' {
			return 0, crex.Wrapf(ErrLex, "non-digit %q in integer %q", string(ch), s)
		}
		next := v*10 + uint64(ch-'0')
		if next < v {
			return 0, crex.Wrapf(ErrLex, "integer %q overflows uint64", s)
		}
		v = next
	}
	return v, nil
}
