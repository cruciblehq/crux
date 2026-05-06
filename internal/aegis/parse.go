package aegis

import (
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/internal/crex"
)

// Recursive-descent parser over a token stream produced by the lexer.
//
// The parser checks that the token stream is structurally well-formed
// according to the grammar, and produces an AST that directly reflects
// the token stream. Subsystems are responsible for semantic validation.
type parser struct {
	tokens []Token // Token stream produced by the lexer.
	pos    int     // Index of the next unconsumed token.
}

// Parses a single grant string and returns its AST representation.
//
// The input must be a single grant of the form ".subsystem name positional*
// kwarg* (where expr)?". The first positional after the subsystem name is
// required; subsequent positionals and any key=value kwargs are interpreted
// by the subsystem builder. No validation of subsystem names, positional
// semantics, or kwarg keys is performed; that is the responsibility of each
// subsystem builder. The parser is position-blind and returns errors without
// any source location; the caller is responsible for stamping a line number
// on any error it surfaces.
func Parse(src string) (*Model, error) {
	tokens, err := Lex(src)
	if err != nil {
		return nil, err
	}
	p := &parser{tokens: tokens}
	return p.parseGrant()
}

// Parses the full grant.
//
// The first token must be a TokenSubsystem (the leading '.' fused with the
// subsystem name), then zero or more positional arguments, zero or more
// key=value kwargs, and an optional where clause. The first positional
// argument must be a NAME token, since it's used as the primary selector to
// indicate the grant entity within the subsystem (a syscall, a capability,
// a cgroup knob, etc.); subsequent arguments are interpreted by the subsystem
// builder. The where clause is parsed as an expression and attached to the AST.
// The parser produces exactly one Grant per source string. If the token stream
// is not well-formed according to the grammar, an error is returned with detail
// about the first encountered problem.
func (p *parser) parseGrant() (*Model, error) {
	subsysTok, err := p.expect(TokenSubsystem)
	if err != nil {
		return nil, crex.Wrapf(ErrParse, "expected %q at the start of a grant", ".subsystem")
	}

	args, err := p.parsePositionalList()
	if err != nil {
		return nil, err
	}

	kwargs, err := p.parseKwargList()
	if err != nil {
		return nil, err
	}

	grant := &Model{
		Subsystem: subsysTok.Text,
		Args:      args,
		Kwargs:    kwargs,
	}

	if p.peek().Type == TokenWhere {
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		grant.Where = expr
	}

	if p.peek().Type != TokenEOF {
		return nil, crex.Wrapf(ErrParse, "unexpected %q after grant", p.peek().Type)
	}

	return grant, nil
}

// Collects one or more positional arguments.
//
// The first positional is required and must be a NAME; any other token in that
// position is an error, as is a NAME immediately followed by an operator (which
// would make it a kwarg instead). Subsequent positionals are consumed until the
// current token is no longer a positional.
func (p *parser) parsePositionalList() ([]Arg, error) {
	headTok := p.peek()
	if headTok.Type != TokenName {
		return nil, crex.Wrapf(ErrParse, "expected name after subsystem, found %q instead", headTok.Type)
	}
	var args []Arg
	for p.isPositional() {
		arg, err := p.parsePositional()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	if len(args) == 0 {
		return nil, crex.Wrapf(ErrParse, "expected positional argument after subsystem, found kwarg %q instead", headTok.Text)
	}
	return args, nil
}

// Parses a single positional argument (NAME, INT, or STRING).
//
// Does not consume tokens beyond the single argument; kwarg detection is
// the caller's responsibility. Returns an error if the current token is
// not a valid argument scalar.
func (p *parser) parsePositional() (Arg, error) {
	tok := p.peek()
	switch tok.Type {
	case TokenName, TokenInt, TokenQuantity, TokenString, TokenVar:
		p.advance()
		return tokenToArg(tok)
	default:
		return Arg{}, crex.Wrapf(ErrParse, "expected argument, found %q instead", tok.Type)
	}
}

// Collects keyword arguments until no more kwargs are present.
//
// A kwarg is detected by looking for a NAME token followed by '='. If the
// current token is not a NAME or is a NAME not followed by '=', then no more
// kwargs are present and the list is complete. Returns an error if any kwarg
// is malformed.
func (p *parser) parseKwargList() ([]Kwarg, error) {
	var kwargs []Kwarg
	for p.isKwarg() {
		kw, err := p.parseKwarg()
		if err != nil {
			return nil, err
		}
		kwargs = append(kwargs, kw)
	}
	return kwargs, nil
}

// Parses a single kwarg.
//
// The key must be a NAME. The '=' separator is required. The right-hand side
// is parsed as a single positional scalar.
func (p *parser) parseKwarg() (Kwarg, error) {
	keyTok := p.peek()
	if keyTok.Type != TokenName {
		return Kwarg{}, crex.Wrapf(ErrParse, "kwarg key must be a name, found %q instead", keyTok.Type)
	}
	p.advance()
	if _, err := p.expect(TokenEq); err != nil {
		return Kwarg{}, crex.Wrapf(ErrParse, "expected %q after kwarg name %q", "=", keyTok.Text)
	}
	val, err := p.parsePositional()
	if err != nil {
		return Kwarg{}, err
	}
	return Kwarg{
		Key:   keyTok.Text,
		Value: val,
	}, nil
}

// Entry point for expression parsing.
//
// Delegates directly to the or-expression level, which has the lowest
// operator precedence.
func (p *parser) parseExpr() (Expr, error) {
	return p.parseOr()
}

// Parses a sequence of and-expressions joined by "or" operators.
//
// Left-associative; produces a chain of BinaryExpr nodes with OpOr. The
// and-expressions themselves may be joined by "and" operators, which are
// parsed at the next level down and have higher precedence.
func (p *parser) parseOr() (Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == TokenOr {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: OpOr, Left: left, Right: right}
	}
	return left, nil
}

// Parses a sequence of not-expressions joined by "and" operators.
//
// Left-associative; produces a chain of BinaryExpr nodes with OpAnd. The
// not-expressions may be prefixed by "not" operators, which are parsed at
// the next level down and have higher precedence.
func (p *parser) parseAnd() (Expr, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == TokenAnd {
		p.advance()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: OpAnd, Left: left, Right: right}
	}
	return left, nil
}

// not-expr = "not" not-expr | primary
//
// "not" here is a unary logical negation, consumed before any field is parsed.
// The two-token "not like" form is handled inside parsePredicate after the LHS
// field has been consumed.
func (p *parser) parseNot() (Expr, error) {
	if p.peek().Type == TokenNot {
		p.advance()
		operand, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: OpNot, Operand: operand}, nil
	}
	return p.parsePrimary()
}

// Parses a parenthesised sub-expression or a predicate.
//
// A leading "(" triggers recursive expression parsing with a required closing
// ")". Any other token is forwarded to parsePredicate.
func (p *parser) parsePrimary() (Expr, error) {
	if p.peek().Type == TokenLParen {
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
		return expr, nil
	}
	return p.parsePredicate()
}

// Parses a predicate expression beginning with a field or value operand.
//
// The operator following the operand determines which form is parsed: a
// relational comparison, set membership, pattern match, range test, or
// bitwise mask test. Returns an error if no recognised operator follows.
func (p *parser) parsePredicate() (Expr, error) {
	field, err := p.parseOperand()
	if err != nil {
		return nil, err
	}
	switch p.peek().Type {
	case TokenEq, TokenNeq, TokenGt, TokenGte, TokenLt, TokenLte:
		return p.parseCompare(field)
	case TokenIn:
		return p.parseIn(field)
	case TokenLike:
		p.advance()
		return p.parseLikePattern(field, false)
	case TokenNot:
		return p.parseNotLike(field)
	case TokenBetween:
		return p.parseBetween(field)
	case TokenAmpersand:
		return p.parseBitTest(field)
	default:
		return nil, crex.Wrapf(ErrParse, "expected operator after %s, found %q instead", field, p.peek().Type)
	}
}

// Parses a comparison expression.
//
// The syntax is "field op value" where op is an operator recognized by
// isOperator. Both field and value are parsed as operands, so they may
// be field references, integer literals, or string literals. Returns an
// error if the operator is not recognised or if the right-hand value is
// not a valid operand.
func (p *parser) parseCompare(field Operand) (Expr, error) {
	op := p.advance()
	cmpOp, err := tokenToCmpOp(op.Type)
	if err != nil {
		return nil, crex.Wrapf(ErrParse, "unexpected comparison operator %q", op.Type)
	}
	right, err := p.parseOperand()
	if err != nil {
		return nil, err
	}
	return &CompareExpr{Left: field, Op: cmpOp, Right: right}, nil
}

// Parses in-expressions.
//
// The syntax is "field in (value, value, ...)". Returns an error if the "in"
// operator is not present or if the parenthesised value list is malformed.
func (p *parser) parseIn(field Operand) (Expr, error) {
	if _, err := p.expect(TokenIn); err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenLParen); err != nil {
		return nil, crex.Wrapf(ErrParse, "expected %q after %q", "(", "in")
	}
	var values []Operand
	for {
		val, err := p.parseOperand()
		if err != nil {
			return nil, err
		}
		values = append(values, val)
		if p.peek().Type != TokenComma {
			break
		}
		p.advance()
	}
	if _, err := p.expect(TokenRParen); err != nil {
		return nil, crex.Wrapf(ErrParse, "expected %q after %q list", ")", "in")
	}
	return &InExpr{Field: field, Values: values}, nil
}

// Parses like-pattern expressions.
//
// The syntax is "field like STRING" or "field not like STRING". The pattern is
// a string literal whose content is not interpreted by the parser. Returns an
// error if the operator is not present or if the pattern is not a string literal.
func (p *parser) parseLikePattern(field Operand, negated bool) (Expr, error) {
	patTok := p.peek()
	if patTok.Type != TokenString {
		if negated {
			return nil, crex.Wrapf(ErrParse, "expected pattern after %q", "not like")
		}
		return nil, crex.Wrapf(ErrParse, "expected pattern after %q", "like")
	}
	p.advance()
	pat, err := patTok.Unquote()
	if err != nil {
		return nil, crex.Wrapf(ErrParse, "invalid pattern: %w", err)
	}
	return &LikeExpr{Field: field, Pattern: pat, Negated: negated}, nil
}

// Parses negated like-pattern expressions.
//
// The syntax is "field not like STRING". The pattern is a string literal whose
// content is not interpreted by the parser. Returns an error if the "not like"
// operator is not present or if the pattern is not a string literal.
func (p *parser) parseNotLike(field Operand) (Expr, error) {
	if p.peekN(1).Type != TokenLike {
		return nil, crex.Wrapf(ErrParse, "expected %q after %q in predicate", "like", "not")
	}
	if _, err := p.expect(TokenNot); err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenLike); err != nil {
		return nil, err
	}
	return p.parseLikePattern(field, true)
}

// Parses between-expressions.
//
// The syntax is "field between low and high". Returns an error if the
// "between" operator is not present or if the "and" keyword is missing
// between the low and high operands.
func (p *parser) parseBetween(field Operand) (Expr, error) {
	if _, err := p.expect(TokenBetween); err != nil {
		return nil, err
	}
	low, err := p.parseOperand()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenAnd); err != nil {
		return nil, crex.Wrapf(ErrParse, "expected %q in between expression", "and")
	}
	high, err := p.parseOperand()
	if err != nil {
		return nil, err
	}
	return &BetweenExpr{Field: field, Low: low, High: high}, nil
}

// Parses bit-test expressions.
//
// The syntax is "field & mask [= value]". Returns an error if the "&" operator
// is not present or if the optional "=" operator is present without a value.
func (p *parser) parseBitTest(field Operand) (Expr, error) {
	if _, err := p.expect(TokenAmpersand); err != nil {
		return nil, err
	}
	mask, err := p.parseOperand()
	if err != nil {
		return nil, err
	}
	var val *Operand
	if p.peek().Type == TokenEq {
		p.advance()
		v, err := p.parseOperand()
		if err != nil {
			return nil, err
		}
		val = &v
	}
	return &BitTestExpr{Field: field, Mask: mask, Val: val}, nil
}

// Parses a field reference, integer literal, or string literal.
//
// Field references are TokenName tokens that may contain dots for dotted names
// like "file.path". Both sides of a comparison may be field references, (e.g.,
// expressions like "task.uid = target.uid").
func (p *parser) parseOperand() (Operand, error) {
	tok := p.peek()
	switch tok.Type {
	case TokenName:
		p.advance()
		return Operand{IsField: true, Field: tok.Text}, nil
	case TokenInt:
		p.advance()
		v, err := parseIntLiteral(tok.Text)
		if err != nil {
			return Operand{}, crex.Wrapf(ErrParse, "invalid integer literal %q: %w", tok.Text, err)
		}
		return Operand{Value: Value{Type: ValueInt, Int: v}}, nil
	case TokenQuantity:
		return Operand{}, crex.Wrapf(ErrParse, "quantity literals are not yet supported in expressions")
	case TokenString:
		p.advance()
		decoded, err := tok.Unquote()
		if err != nil {
			return Operand{}, crex.Wrapf(ErrParse, "invalid string literal: %w", err)
		}
		enc, err := tokEncodingToAST(tok.Encoding)
		if err != nil {
			return Operand{}, crex.Wrapf(ErrParse, "invalid string encoding: %w", err)
		}
		return Operand{
			Value: Value{
				Type:        ValueStr,
				Str:         decoded,
				StrEncoding: enc,
			},
		}, nil
	case TokenVar:
		p.advance()
		return Operand{Value: Value{Type: ValueVar, Str: tok.Text}}, nil
	default:
		return Operand{}, crex.Wrapf(ErrParse, "expected field or value, found %q instead", tok.Type)
	}
}

// Parses a C-style integer literal string produced by the lexer into a uint64.
//
// Recognises all bases: unprefixed decimal, 0x/0X hexadecimal, 0o/0O octal,
// 0b/0B binary, and legacy leading-0 octal.
func parseIntLiteral(s string) (uint64, error) {
	lower := strings.ToLower(s)
	switch {
	case strings.HasPrefix(lower, "0x"):
		return strconv.ParseUint(s[2:], 16, 64)
	case strings.HasPrefix(lower, "0o"):
		return strconv.ParseUint(s[2:], 8, 64)
	case strings.HasPrefix(lower, "0b"):
		return strconv.ParseUint(s[2:], 2, 64)
	case len(s) > 1 && s[0] == '0':
		return strconv.ParseUint(s[1:], 8, 64)
	default:
		return strconv.ParseUint(s, 10, 64)
	}
}

// Returns the current token without advancing.
//
// Returns the token at the current position, or a TokenEOF if the position is
// past the end of the stream, so that callers do not need to bounds-check.
func (p *parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

// Returns the token n positions ahead without advancing.
//
// Like peek, except the returned token is from the future position. Returns
// a TokenEOF when the requested position is past the end of the stream.
func (p *parser) peekN(n int) Token {
	idx := p.pos + n
	if idx >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[idx]
}

// Consumes and returns the current token.
//
// If the current position is past the end of the stream, returns a TokenEOF
// without advancing, so callers can drain the stream without bounds-checking.
func (p *parser) advance() Token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

// Consumes the current token if it matches typ, otherwise returns an error.
//
// Returns the consumed token and nil on success. On a type mismatch, returns
// a zero token and a parse error.
func (p *parser) expect(typ TokenType) (Token, error) {
	tok := p.peek()
	if tok.Type != typ {
		return Token{}, crex.Wrapf(ErrParse, "expected %q, found %q instead", typ, tok.Type)
	}
	return p.advance(), nil
}

// Maps a comparison-operator token type to its AST counterpart.
//
// Returns an error for any token type that is not one of the six recognised
// comparison operators.
func tokenToCmpOp(t TokenType) (CmpOp, error) {
	switch t {
	case TokenEq:
		return CmpEq, nil
	case TokenNeq:
		return CmpNeq, nil
	case TokenGt:
		return CmpGt, nil
	case TokenGte:
		return CmpGte, nil
	case TokenLt:
		return CmpLt, nil
	case TokenLte:
		return CmpLte, nil
	default:
		return "", crex.Wrapf(ErrParse, "not a comparison operator: %q", t)
	}
}

// Converts a token to a positional argument, decoding string literals.
//
// Handles NAME, INT, QUANTITY, STRING, and VAR tokens. For STRING tokens the
// surrounding quotes are stripped and recognised escape sequences are resolved
// via Token.Unquote.
func tokenToArg(t Token) (Arg, error) {
	switch t.Type {
	case TokenName:
		return Arg{Type: ArgName, Value: t.Text}, nil
	case TokenInt:
		return Arg{Type: ArgInt, Value: t.Text}, nil
	case TokenQuantity:
		return Arg{Type: ArgQuantity, Value: t.Text}, nil
	case TokenString:
		unquoted, err := t.Unquote()
		if err != nil {
			return Arg{}, crex.Wrapf(ErrParse, "invalid string literal: %w", err)
		}
		enc, err := tokEncodingToAST(t.Encoding)
		if err != nil {
			return Arg{}, crex.Wrapf(ErrParse, "invalid string encoding: %w", err)
		}
		argType := ArgStrASCII
		if enc == StrUnicode {
			argType = ArgStrUnicode
		}
		return Arg{Type: argType, Value: unquoted}, nil
	case TokenVar:
		return Arg{Type: ArgVar, Value: t.Text}, nil
	default:
		return Arg{}, crex.Wrapf(ErrParse, "unsupported argument token %q", t.Type)
	}
}

// Maps a lexer string-token encoding to its AST counterpart.
//
// Returns an error for any encoding value that is not explicitly recognised.
func tokEncodingToAST(enc TokenEncoding) (StrEncoding, error) {
	switch enc {
	case EncodingASCII:
		return StrASCII, nil
	case EncodingUnicode:
		return StrUnicode, nil
	default:
		return 0, crex.Wrapf(ErrParse, "unknown token encoding %q", enc)
	}
}

// Whether the current token is a positional argument.
//
// A positional is a NAME, INT, STRING, or VAR token that is not immediately
// followed by an operator. A value-shaped token followed by an operator is a
// kwarg attempt and is left for parseKwargList to consume so that bad-key
// and bad-operator cases can be reported with a precise error.
func (p *parser) isPositional() bool {
	switch p.peek().Type {
	case TokenName, TokenInt, TokenQuantity, TokenString, TokenVar:
		return !p.peekN(1).Type.isOperator()
	default:
		return false
	}
}

// Whether the current token starts a kwarg.
//
// Returns true for any value-shaped token (NAME, INT, STRING, VAR) followed
// by an operator. The kwarg parser is responsible for validating that the
// key is a NAME and the operator is '='; this helper deliberately accepts
// any operator so that bad-key and bad-operator cases reach the parser and
// produce a precise error rather than being silently reinterpreted.
func (p *parser) isKwarg() bool {
	switch p.peek().Type {
	case TokenName, TokenInt, TokenQuantity, TokenString, TokenVar:
		return p.peekN(1).Type.isOperator()
	default:
		return false
	}
}
