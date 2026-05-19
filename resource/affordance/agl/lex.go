package agl

import (
	"fmt"
	"unicode/utf8"

	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/resource/affordance/units"
)

// Lexer for AGL grant strings.
//
// The lexer produces a stream of tokens with type, text, and encoding. The
// tokens are designed to be consumed by the parser, which validates the token
// sequence and constructs an AST. The tokenization can be initiated with Lex.
// It is not reusable. The entire input is consumed in a single pass, and the
// first lex error is fatal. The lexer handles single-line input only and
// produces no positional information; the caller is responsible for stamping
// a line number on any error it surfaces.
type lexer struct {
	src    string  // Source text being lexed.
	pos    int     // Current byte position in src.
	tokens []Token // Accumulated tokens.
	err    error   // First error encountered, or nil.
}

// Tokenises the entire source string.
//
// Tokens are accumulated until EOF or an error occurs. All tokens are emitted
// with type, text, and encoding. The encoding is EncodingNone except for string
// tokens, which are EncodingASCII for quoted ASCII literals and unquoted globs,
// or EncodingUnicode for quoted literals prefixed with 'u'. Null bytes produce
// a lex error wherever they occur in the input. The returned slice always ends
// with TokenEOF. If an error is encountered the slice also contains a TokenError
// at the point of failure, and the returned error is non-nil. Returned errors
// do not include positional information.
func Lex(src string) ([]Token, error) {
	l := &lexer{src: src}
	l.run()
	return l.tokens, l.err
}

// Consumes the input and emits tokens until EOF or an error is reached.
//
// Skips leading whitespace before each token. When EOF is reached, a TokenEOF
// is emitted. If any lex error occurs, a TokenError is emitted at the point
// of failure and the error is recorded in l.err. Once an error is recorded,
// no further tokens are emitted (other than TokenEOF).
func (l *lexer) run() {
	for l.more() && l.err == nil {
		l.skipWhitespace()
		if !l.more() {
			break
		}
		l.lexToken()
	}
	l.emit(TokenEOF, "")
}

// Consumes one token.
//
// A leading '.' starts a subsystem header and is fused with the following
// name into a single TokenSubsystem; whitespace between the dot and the name
// is not permitted. Punctuation and operators map to their respective token
// types. A " indicates a string literal, which may be prefixed with u for
// Unicode. An unquoted glob beginning with '/' is consumed as a string token
// with EncodingASCII. A name is any sequence of letters, digits, underscores,
// dots, or internal hyphens that begins with a letter or underscore and is
// not a reserved keyword; a hyphen joins two name characters only when the
// characters on both sides are themselves name characters, so trailing or
// surrounded-by-whitespace hyphens still tokenise as TokenMinus. Numbers are
// consumed as unsigned integers in C-style bases (decimal, 0x hex, 0o/0
// octal, 0b binary); a plain decimal integer immediately followed by one of
// the recognised quantity suffixes (Ki, Mi, Gi, Ti, Pi, Ei, k, K, M, G, T,
// P, E, m, u, n) and a clean terminator fuses into a single TokenQuantity.
// Keywords are case-sensitive (lower-case only). Any unrecognised character
// produces a lex error, including null. The source position of the emitted
// token corresponds to the first character of the token in the source string.
func (l *lexer) lexToken() {
	ch, la := l.peek(), l.lookahead()

	switch {
	case ch == '.':
		l.lexSubsystem()
	case ch == '(':
		l.lexSingle(ch, TokenLParen)
	case ch == ')':
		l.lexSingle(ch, TokenRParen)
	case ch == ',':
		l.lexSingle(ch, TokenComma)
	case ch == '&':
		l.lexSingle(ch, TokenAmpersand)
	case ch == '-':
		l.lexSingle(ch, TokenMinus)
	case ch == '=':
		l.lexSingle(ch, TokenEq)
	case ch == '!':
		l.lexCompound(ch, '=', TokenError, TokenNeq)
	case ch == '>':
		l.lexCompound(ch, '=', TokenGt, TokenGte)
	case ch == '<':
		l.lexCompound(ch, '=', TokenLt, TokenLte)
	case ch == '"':
		l.lexString()
	case ch == '/':
		l.lexGlob()
	case ch == 'u' && la == '"':
		l.lexUnicodeString()
	case isDigit(ch):
		l.lexNumber()
	case isNameStart(ch):
		l.lexName()
	case ch == '$':
		l.lexVar()
	case ch == 0:
		l.errorf("null byte in input")
	default:
		l.errorf("unexpected character %q", string(ch))
	}
}

// Appends a token with explicit encoding.
//
// The token's type and text are set as given, and the encoding is set to enc.
// The encoding should be EncodingNone for non-string tokens, or EncodingASCII
// or EncodingUnicode for string tokens.
func (l *lexer) emitWithEncoding(typ TokenType, text string, enc TokenEncoding) {
	l.tokens = append(l.tokens, Token{
		Type:     typ,
		Encoding: enc,
		Text:     text,
	})
}

// Appends a token with EncodingNone.
//
// Used for tokens whose encoding is not meaningful (everything except
// TokenString). To specify an encoding, use emitWithEncoding instead.
func (l *lexer) emit(typ TokenType, text string) {
	l.emitWithEncoding(typ, text, EncodingNone)
}

// Consumes a single character.
//
// The emitted token has the given type and the consumed character as its
// Text and EncodingNone since single-character tokens are never strings.
func (l *lexer) lexSingle(ch rune, typ TokenType) {
	l.consume(ch)
	l.emit(typ, string(ch))
}

// Consumes one- or two-character operators with a common prefix.
//
// ch is the required first character and expected is the optional second
// character. If expected follows ch, the compound token type is emitted;
// otherwise the single token type is emitted, or a lex error is reported
// when single is TokenError. The Text is the literal source spelling of the
// operator (e.g. ">" or ">="). The emitted token has EncodingNone since
// operators are never strings.
func (l *lexer) lexCompound(ch, expected rune, single, compound TokenType) {
	if l.lookahead() == expected {
		l.consume(ch, expected)
		l.emit(compound, string([]rune{ch, expected}))
	} else if single == TokenError {
		l.errorf("unexpected character %q", string(ch))
	} else {
		l.consume(ch)
		l.emit(single, string(ch))
	}
}

// Consumes an ASCII string literal.
//
// Consumes the opening quote, body, and closing quote. The emitted Text is
// the verbatim source slice including the surrounding quotes and any escape
// sequences in their unresolved form. Use Unquote to obtain the decoded
// content. The token has type TokenString and EncodingASCII; non-ASCII
// characters in the body are rejected (use lexUnicodeString for Unicode).
// Only \" and \\ are recognised as escape sequences. The opening quote at
// the current position is consumed; if it is not a '"', a lex error is
// recorded. A newline or EOF before the closing quote is an error.
func (l *lexer) lexString() {
	startOff := l.pos
	l.consume('"')
	l.lexStringBody(TokenString, startOff, EncodingASCII)
}

// Consumes a Unicode string literal.
//
// Consumes the 'u' prefix and opening quote before consuming the body and
// closing quote. The emitted Text is the verbatim source slice including
// the 'u' prefix, surrounding quotes, and any escape sequences in their
// unresolved form. Use Unquote to obtain the decoded content. The token
// has type TokenString and EncodingUnicode. Only \" and \\ are recognised
// as escape sequences. The 'u' prefix and opening quote at the current
// position are consumed; if either is missing, a lex error is recorded.
// Non-ASCII characters are permitted. Use lexString for ASCII-only string
// literals. A newline or EOF before the closing quote is an error.
func (l *lexer) lexUnicodeString() {
	startOff := l.pos
	l.consume('u', '"')
	l.lexStringBody(TokenString, startOff, EncodingUnicode)
}

// Consumes a string literal's body with the given type and encoding.
//
// Consumes the body and closing quote of a string literal, assuming the
// opening quote has already been consumed and the current character is the
// first byte of the body. The emitted Text is the verbatim source slice from
// startOff through (and including) the closing quote, with escape sequences
// left unresolved. Only \" and \\ are recognised as escape sequences; an
// unrecognised escape is a lex error. When enc is EncodingASCII, non-ASCII
// characters in the body are rejected. A newline or EOF before the closing
// quote is an error.
func (l *lexer) lexStringBody(typ TokenType, startOff int, enc TokenEncoding) {
	for l.more() {
		ch := l.peek()
		switch {
		case ch == '"':
			l.consume(ch)
			l.emitWithEncoding(typ, l.src[startOff:l.pos], enc)
			return
		case ch == '\\':
			l.lexEscapeChar()
		case ch == '\n':
			l.errorf("unterminated string literal")
			return
		case ch == 0:
			l.errorf("null byte in string literal")
			return
		case enc == EncodingASCII && ch >= utf8.RuneSelf:
			l.errorf("non-ASCII character %q in string literal", string(ch))
			return
		default:
			l.consume(ch)
		}

		if l.err != nil {
			return
		}
	}
	l.errorf("unterminated string literal")
}

// Consumes a string escape sequence.
//
// Only \" and \\ are recognised. The backslash and the escaped character are
// consumed but no value is materialised; decoding is deferred to Unquote. If
// the escape is not recognised, a lex error is recorded. The source position
// corresponds to the backslash character. A newline or EOF after the
// backslash is an error.
func (l *lexer) lexEscapeChar() {
	l.consume('\\')
	switch esc := l.peek(); esc {
	case '"', '\\':
		l.consume(esc)
	case 0:
		l.errorf("unterminated string escape")
	default:
		l.errorf("unknown escape sequence \\%c", esc)
	}
}

// Consumes an unquoted glob and emits a TokenString.
//
// The leading '/' is included in the emitted Text. Only isGlob characters
// are permitted. Whitespace, EOF, and grammar punctuation terminate cleanly.
// Non-ASCII characters and disallowed ASCII characters are a lex error.
// The emitted token has EncodingASCII since unquoted globs are ASCII-only.
func (l *lexer) lexGlob() {
	startOff := l.pos
	for {
		ch := l.peek()
		if ch == 0 {
			break
		}
		if ch >= utf8.RuneSelf {
			l.errorf("non-ASCII character in unquoted glob")
			return
		}
		if isGlobTerminator(ch) {
			break
		}
		if !isGlob(ch) {
			l.errorf("character %q not allowed in unquoted glob", string(ch))
			return
		}
		l.consume(ch)
	}
	l.emitWithEncoding(TokenString, l.src[startOff:l.pos], EncodingASCII)
}

// Consumes a C-style integer literal in any base, optionally fused with a
// quantity suffix.
//
// Supports unprefixed decimals, 0x/0X hex, 0o/0O octal, 0b/0B binary, and
// legacy leading-0 octal (e.g. 0644). Digit separators are not supported.
// A plain decimal integer immediately followed by one of the recognised
// quantity suffixes (see quantitySuffixes) and a clean terminator emits a
// single TokenQuantity covering both the digits and the suffix; otherwise a
// TokenInt is emitted for the digits alone. Base-prefixed and legacy octal
// literals never fuse with a suffix.
func (l *lexer) lexNumber() {
	start := l.pos
	if l.peek() == '0' && !isQuantityCandidate(l.lookahead()) {
		l.lexZeroPrefixedInt()
		if l.err == nil {
			l.emit(TokenInt, l.src[start:l.pos])
		}
		return
	}

	l.lexDecimalInt()
	if n := l.matchQuantitySuffix(); n > 0 {
		for i := 0; i < n; i++ {
			l.skip()
		}
		l.emit(TokenQuantity, l.src[start:l.pos])
		return
	}
	l.emit(TokenInt, l.src[start:l.pos])
}

// Whether ch could begin a quantity suffix.
//
// Used to distinguish a bare-zero quantity (e.g. "0Gi") from a base-prefixed
// or legacy-octal literal. Returns true for any ASCII letter that may start a
// recognised suffix; the suffix itself is validated by matchQuantitySuffix.
func isQuantityCandidate(ch rune) bool {
	switch ch {
	case 'K', 'M', 'G', 'T', 'P', 'E', 'k', 'm', 'u', 'n':
		return true
	}
	return false
}

// Returns the byte length of a recognised quantity suffix that begins at the
// current position, or 0 when no suffix matches.
//
// Scans an unbroken run of ASCII letters at the current position without
// consuming, looks the run up in quantitySuffixes, and confirms that the
// character immediately following the run is not itself a name continuation
// character (so longer identifiers like "Gigabyte" are not mistaken for the
// suffix "Gi"). Returns 0 when the run is empty, unknown, or run into more
// name characters.
func (l *lexer) matchQuantitySuffix() int {
	save := l.pos
	for {
		ch := l.peek()
		if !isLetter(ch) {
			break
		}
		l.skip()
	}
	end := l.pos
	cand := l.src[save:end]
	l.pos = save
	if cand == "" {
		return 0
	}
	if !units.IsKnown(cand) {
		return 0
	}
	next, _ := utf8.DecodeRuneInString(l.src[end:])
	if isNameContinue(next) || next == '-' {
		return 0
	}
	return end - save
}

// Consumes a sequence of decimal digits.
//
// Reads zero or more characters for which isDigit returns true. Any other
// character (including EOF) terminates the literal cleanly without error.
func (l *lexer) lexDecimalInt() {
	for isDigit(l.peek()) {
		l.consume()
	}
}

// Lexes a 0-prefixed integer literal.
//
// The leading '0' must be the current character. A lone '0' not followed by a
// base prefix or octal digit is a valid decimal zero. The digits 8 and 9 after
// a bare '0' are always an error.
func (l *lexer) lexZeroPrefixedInt() {
	next := l.lookahead()
	switch {
	case next == 'x' || next == 'X':
		l.lexPrefixedInt(isHexDigit, "hex digit after 0x")
	case next == 'o' || next == 'O':
		l.lexPrefixedInt(isOctalDigit, "octal digit after 0o")
	case next == 'b' || next == 'B':
		l.lexPrefixedInt(isBinaryDigit, "binary digit after 0b")
	case next >= '0' && next <= '7':
		l.lexLegacyOctal()
	case next == '8' || next == '9':
		l.errorf("invalid digit %q in octal literal", string(next))
	default:
		l.consume('0') // Decimal zero.
	}
}

// Lexes the prefix and digits of a base-prefixed integer literal.
//
// The leading '0' must be the current character and the base letter (x/X,
// o/O, or b/B) must follow. '0' is consumed assertively and the base letter
// is skipped unconditionally. isValidDigit reports whether a rune is a valid
// digit for the base. errMsg is the description used in the lex error when
// no digit follows the prefix.
func (l *lexer) lexPrefixedInt(isValidDigit func(rune) bool, errMsg string) {
	l.consume('0')
	l.skip() // Base letter

	if !isValidDigit(l.peek()) {
		l.errorf("expected %s", errMsg)
		return
	}

	for isValidDigit(l.peek()) {
		l.consume()
	}
}

// Lexes a legacy leading-zero octal literal (e.g. 0644).
//
// The leading '0' must be the current character. The digits 8 and 9 are
// always an error.
func (l *lexer) lexLegacyOctal() {
	l.consume('0')
	for isOctalDigit(l.peek()) {
		l.consume()
	}
	if isDigit(l.peek()) {
		l.errorf("invalid digit %q in octal literal", string(l.peek()))
	}
}

// Lexes a name or keyword.
//
// Keywords are case-sensitive; only lower-case forms are recognised (e.g.
// "where", "and"). Dotted names are emitted as a single token. A hyphen is
// folded into the name only when it sits between two name-continuation
// characters; a hyphen at the end of a name run, or one followed by
// whitespace or punctuation, terminates the name and remains available as a
// TokenMinus.
func (l *lexer) lexName() {
	start := l.pos
	for {
		ch := l.peek()
		if isNameContinue(ch) {
			l.consume()
			continue
		}
		if ch == '-' && isNameContinue(l.lookahead()) {
			l.consume()
			continue
		}
		break
	}
	word := l.src[start:l.pos]

	if isKeyword(word) {
		l.emit(TokenType(word), word)
	} else {
		l.emit(TokenName, word)
	}
}

// Consumes a subsystem header of the form ".name".
//
// The leading '.' must be the current character and the name that follows must
// begin immediately, without whitespace. Both are consumed. The emitted token
// has type TokenSubsystem and Text holds the subsystem name without the leading
// '.'. A bare '.' or one followed by whitespace or a non-name character is a
// lex error.
func (l *lexer) lexSubsystem() {
	l.consume('.')
	if !isNameStart(l.peek()) {
		l.errorf("expected subsystem name after %q", ".")
		return
	}
	start := l.pos
	for isVarContinue(l.peek()) {
		l.consume()
	}
	l.emit(TokenSubsystem, l.src[start:l.pos])
}

// Consumes a variable reference of the form "$name".
//
// The leading '$' must be the current character and is consumed. The name
// that follows must begin with a name-start character and continues with
// name-continue characters; dots are not permitted in variable names. The
// emitted token has type TokenVar and Text holds the variable name without
// the '$' prefix. A bare '$' or '$' followed by a non-name character is a
// lex error.
func (l *lexer) lexVar() {
	l.consume('$')
	if !isNameStart(l.peek()) {
		l.errorf("expected variable name after %q", "$")
		return
	}
	start := l.pos
	for isVarContinue(l.peek()) {
		l.consume()
	}
	l.emit(TokenVar, l.src[start:l.pos])
}

// Skips whitespace.
func (l *lexer) skipWhitespace() {
	for isWhitespace(l.peek()) {
		l.consume()
	}
}

// Returns the rune at the current position without advancing.
//
// Returns 0 if there is no current character. Null bytes (0) are rejected as
// invalid input, so 0 is unambiguously the EOF sentinel.
func (l *lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.src[l.pos:])
	return r
}

// Returns the rune immediately after the current position without advancing.
//
// The next position is computed by skipping the full byte width of the current
// rune, so this is safe even when the current character is multi-byte. Returns
// 0 if there is no next character. Null bytes (0) are rejected as invalid input
// by the lexer, so 0 is unambiguously the EOF sentinel.
func (l *lexer) lookahead() rune {
	_, size := utf8.DecodeRuneInString(l.src[l.pos:])
	next := l.pos + size
	if next >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.src[next:])
	return r
}

// Advances past one or more runes.
//
// With no arguments, unconditionally consumes the current rune. With one or
// more expected runes, each is verified against the current position before
// advancing; a mismatch records a lex error and returns 0 without consuming
// further. Returns the rune now at the current position after the last
// consumed character, or 0 at EOF.
func (l *lexer) consume(expected ...rune) rune {
	if len(expected) == 0 {
		if !l.more() {
			return 0
		}
		l.skip()
		return l.peek()
	}
	for _, want := range expected {
		if !l.more() {
			l.errorf("unexpected EOF, expected %q", want)
			return 0
		}
		if got := l.peek(); got != want {
			l.errorf("unexpected %q, expected %q", got, want)
			return 0
		}
		l.skip()
	}
	return l.peek()
}

// Advances one rune.
//
// pos is advanced by the full byte width of the current rune.
func (l *lexer) skip() {
	_, size := utf8.DecodeRuneInString(l.src[l.pos:])
	l.pos += size
}

// Sets l.err and emits a TokenError carrying the message text.
//
// The error message should not include positional information; the lexer is
// position-blind and the caller is responsible for stamping a line number on
// any error it surfaces.
func (l *lexer) errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.err = crex.Wrapf(ErrLex, "%s", msg)
	l.tokens = append(l.tokens, Token{
		Type: TokenError,
		Text: msg,
	})
}

// Whether there is more input to consume.
func (l *lexer) more() bool {
	return l.pos < len(l.src)
}
