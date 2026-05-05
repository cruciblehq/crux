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
