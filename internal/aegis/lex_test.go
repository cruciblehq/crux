package aegis

import (
	"testing"
)

// Returns the first token from src.
func lexOne(t *testing.T, src string) Token {
	t.Helper()
	toks, err := Lex(src)
	if err != nil {
		t.Fatalf("Lex(%q): unexpected error: %v", src, err)
	}
	if len(toks) == 0 {
		t.Fatalf("Lex(%q): no tokens", src)
	}
	return toks[0]
}

// Returns all non-EOF tokens from src.
func lexTokens(t *testing.T, src string) []Token {
	t.Helper()
	toks, err := Lex(src)
	if err != nil {
		t.Fatalf("Lex(%q): unexpected error: %v", src, err)
	}
	var out []Token
	for _, tok := range toks {
		if tok.Type != TokenEOF {
			out = append(out, tok)
		}
	}
	return out
}

func TestEmptyInput(t *testing.T) {
	toks, err := Lex("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(toks) != 1 || toks[0].Type != TokenEOF {
		t.Fatalf("expected [EOF], got %v", toks)
	}
}

func TestWhitespaceOnly(t *testing.T) {
	toks, err := Lex("   \t  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(toks) != 1 || toks[0].Type != TokenEOF {
		t.Fatalf("expected [EOF], got %v", toks)
	}
}

func TestEOFAlwaysLast(t *testing.T) {
	for _, src := range []string{"", "foo", "42", `"hello"`, "/usr/bin", "where"} {
		t.Run(src, func(t *testing.T) {
			toks, err := Lex(src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(toks) == 0 || toks[len(toks)-1].Type != TokenEOF {
				t.Fatalf("last token is not EOF: %v", toks)
			}
		})
	}
}

func TestNullByteInInput(t *testing.T) {
	_, err := Lex("foo\x00bar")
	if err == nil {
		t.Fatal("expected error for null byte in input, got nil")
	}
}

func TestUnrecognisedCharacters(t *testing.T) {
	for _, src := range []string{"@", "#", "$", "%", "^", "~", "`", ";", "\n", "\r"} {
		t.Run(src, func(t *testing.T) {
			_, err := Lex(src)
			if err == nil {
				t.Errorf("Lex(%q): expected error, got nil", src)
			}
		})
	}
}

func TestErrorTokenEmitted(t *testing.T) {
	toks, err := Lex("@")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	for _, tok := range toks {
		if tok.Type == TokenError {
			return
		}
	}
	t.Error("no TokenError in slice after lex error")
}

func TestErrorStopsLexing(t *testing.T) {
	// After the error on '@', no further non-EOF tokens should be emitted.
	toks, err := Lex("foo @ bar")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	for _, tok := range toks {
		if tok.Type == TokenName && tok.Text == "bar" {
			t.Error("lexing continued past error: found 'bar' token")
		}
	}
}

func TestConsumeExpectedAtEOF(t *testing.T) {
	l := &lexer{src: ""}
	l.consume('x')
	if l.err == nil {
		t.Error("expected error when consuming expected rune at EOF, got nil")
	}
}

func TestVarToken(t *testing.T) {
	cases := []struct {
		src  string
		want string
	}{
		{"$x", "x"},
		{"$cpumin", "cpumin"},
		{"$_foo", "_foo"},
		{"$a1_b2", "a1_b2"},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			tok := lexOne(t, tc.src)
			if tok.Type != TokenVar {
				t.Errorf("Type: got %s, want %s", tok.Type, TokenVar)
			}
			if tok.Text != tc.want {
				t.Errorf("Text: got %q, want %q", tok.Text, tc.want)
			}
		})
	}
}

func TestVarTokenErrors(t *testing.T) {
	cases := []string{
		"$",
		"$1",
		"$.foo",
		"$ ",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			_, err := Lex(src)
			if err == nil {
				t.Errorf("Lex(%q): expected error, got nil", src)
			}
		})
	}
}

func TestVarNoDot(t *testing.T) {
	// A '.' after a variable name terminates the variable; the dot then
	// fuses with the following name into a TokenSubsystem.
	toks, err := Lex("$a.b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(toks) < 3 {
		t.Fatalf("expected at least 3 tokens, got %d", len(toks))
	}
	if toks[0].Type != TokenVar || toks[0].Text != "a" {
		t.Errorf("token 0: got %s %q, want var \"a\"", toks[0].Type, toks[0].Text)
	}
	if toks[1].Type != TokenSubsystem || toks[1].Text != "b" {
		t.Errorf("token 1: got %s %q, want subsystem \"b\"", toks[1].Type, toks[1].Text)
	}
}

func TestConsumeExpectedMismatch(t *testing.T) {
	l := &lexer{src: "a"}
	l.consume('b')
	if l.err == nil {
		t.Error("expected error when consumed rune does not match expected, got nil")
	}
	if l.pos != 0 {
		t.Errorf("pos must not advance on mismatch: got %d, want 0", l.pos)
	}
}

func TestConsumeUnconditionalAtEOF(t *testing.T) {
	l := &lexer{src: ""}
	got := l.consume()
	if got != 0 {
		t.Errorf("consume() at EOF: got %q, want 0", got)
	}
	if l.err != nil {
		t.Errorf("consume() at EOF must not set err, got: %v", l.err)
	}
	if l.pos != 0 {
		t.Errorf("consume() at EOF must not advance pos: got %d, want 0", l.pos)
	}
}

func TestOperators(t *testing.T) {
	cases := []struct {
		src string
		typ TokenType
	}{
		{"(", TokenLParen},
		{")", TokenRParen},
		{",", TokenComma},
		{"&", TokenAmpersand},
		{"-", TokenMinus},
		{"=", TokenEq},
		{"!=", TokenNeq},
		{">", TokenGt},
		{">=", TokenGte},
		{"<", TokenLt},
		{"<=", TokenLte},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			tok := lexOne(t, tc.src)
			if tok.Type != tc.typ {
				t.Errorf("Type: got %s, want %s", tok.Type, tc.typ)
			}
			if tok.Text != tc.src {
				t.Errorf("Text: got %q, want %q", tok.Text, tc.src)
			}
		})
	}
}

func TestBareExclamation(t *testing.T) {
	for _, src := range []string{"!", "!foo", "! ="} {
		t.Run(src, func(t *testing.T) {
			_, err := Lex(src)
			if err == nil {
				t.Errorf("Lex(%q): expected error, got nil", src)
			}
		})
	}
}

func TestName(t *testing.T) {
	cases := []string{
		"foo", "Foo", "FOO",
		"_foo", "foo_bar", "foo_",
		"arg0", "x1y2",
		"file.path", "task.uid",
		"u",    // bare 'u' without a following quote is a name
		"user", // 'u' followed by non-quote is a name
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			tok := lexOne(t, src)
			if tok.Type != TokenName {
				t.Errorf("Type: got %s, want name", tok.Type)
			}
			if tok.Text != src {
				t.Errorf("Text: got %q, want %q", tok.Text, src)
			}
		})
	}
}

func TestKeywordsRecognised(t *testing.T) {
	cases := []struct {
		src string
		typ TokenType
	}{
		{"where", TokenWhere},
		{"and", TokenAnd},
		{"or", TokenOr},
		{"not", TokenNot},
		{"in", TokenIn},
		{"like", TokenLike},
		{"between", TokenBetween},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			tok := lexOne(t, tc.src)
			if tok.Type != tc.typ {
				t.Errorf("Type: got %s, want %s", tok.Type, tc.typ)
			}
			if tok.Text != tc.src {
				t.Errorf("Text: got %q, want %q", tok.Text, tc.src)
			}
		})
	}
}

func TestKeywordsCaseSensitive(t *testing.T) {
	// Upper-case and mixed-case forms must be lexed as names, not keywords.
	cases := []string{
		"WHERE", "AND", "OR", "NOT", "IN", "LIKE", "BETWEEN",
		"Where", "And", "Or", "Not", "In", "Like", "Between",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			tok := lexOne(t, src)
			if tok.Type != TokenName {
				t.Errorf("Lex(%q): Type: got %s, want name", src, tok.Type)
			}
			if tok.Text != src {
				t.Errorf("Lex(%q): Text: got %q, want %q", src, tok.Text, src)
			}
		})
	}
}

func TestIntegerDecimal(t *testing.T) {
	for _, src := range []string{"0", "1", "42", "1048576", "9999999999"} {
		t.Run(src, func(t *testing.T) {
			tok := lexOne(t, src)
			if tok.Type != TokenInt {
				t.Errorf("Type: got %s, want %s", tok.Type, TokenInt)
			}
			if tok.Text != src {
				t.Errorf("Text: got %q, want %q", tok.Text, src)
			}
		})
	}
}

func TestIntegerHex(t *testing.T) {
	// Source text case must be preserved; the lexer must not normalise to lowercase.
	for _, src := range []string{"0xff", "0xFF", "0xABCDEF", "0x0", "0Xff", "0XDEAD"} {
		t.Run(src, func(t *testing.T) {
			tok := lexOne(t, src)
			if tok.Type != TokenInt {
				t.Errorf("Type: got %s, want %s", tok.Type, TokenInt)
			}
			if tok.Text != src {
				t.Errorf("Text: got %q, want %q (case must be preserved)", tok.Text, src)
			}
		})
	}
}

func TestIntegerOctal(t *testing.T) {
	for _, src := range []string{"0o755", "0O644", "0o0", "0O777"} {
		t.Run(src, func(t *testing.T) {
			tok := lexOne(t, src)
			if tok.Type != TokenInt {
				t.Errorf("Type: got %s, want %s", tok.Type, TokenInt)
			}
			if tok.Text != src {
				t.Errorf("Text: got %q, want %q", tok.Text, src)
			}
		})
	}
}

func TestIntegerBinary(t *testing.T) {
	for _, src := range []string{"0b0", "0b1010", "0B1111", "0b10101010"} {
		t.Run(src, func(t *testing.T) {
			tok := lexOne(t, src)
			if tok.Type != TokenInt {
				t.Errorf("Type: got %s, want %s", tok.Type, TokenInt)
			}
			if tok.Text != src {
				t.Errorf("Text: got %q, want %q", tok.Text, src)
			}
		})
	}
}

func TestIntegerLegacyOctal(t *testing.T) {
	for _, src := range []string{"0644", "0777", "0001", "00"} {
		t.Run(src, func(t *testing.T) {
			tok := lexOne(t, src)
			if tok.Type != TokenInt {
				t.Errorf("Type: got %s, want %s", tok.Type, TokenInt)
			}
			if tok.Text != src {
				t.Errorf("Text: got %q, want %q", tok.Text, src)
			}
		})
	}
}

func TestIntegerErrors(t *testing.T) {
	cases := []string{
		"0x",    // hex: no digits after prefix
		"0X",    // hex (uppercase X): no digits after prefix
		"0o",    // octal: no digits after prefix
		"0O",    // octal (uppercase O): no digits after prefix
		"0b",    // binary: no digits after prefix
		"0B",    // binary (uppercase B): no digits after prefix
		"08",    // legacy octal: digit 8 is invalid
		"09",    // legacy octal: digit 9 is invalid
		"089",   // legacy octal: invalid digit in body
		"06448", // legacy octal: valid octal digits followed by invalid digit
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			_, err := Lex(src)
			if err == nil {
				t.Errorf("Lex(%q): expected error, got nil", src)
			}
		})
	}
}

func TestStringContent(t *testing.T) {
	cases := []struct {
		src     string // source input, also the expected Text
		decoded string // expected Unquote result
	}{
		{`""`, ""},
		{`"hello"`, "hello"},
		{`"say \"hi\""`, `say "hi"`},
		{`"a\\b"`, `a\b`},
		{`"ends with \\"`, `ends with \`},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			tok := lexOne(t, tc.src)
			if tok.Type != TokenString {
				t.Fatalf("Type: got %s, want %s", tok.Type, TokenString)
			}
			if tok.Text != tc.src {
				t.Errorf("Text: got %q, want %q", tok.Text, tc.src)
			}
			got, err := tok.Unquote()
			if err != nil {
				t.Fatalf("Unquote: unexpected error: %v", err)
			}
			if got != tc.decoded {
				t.Errorf("Unquote: got %q, want %q", got, tc.decoded)
			}
		})
	}
}

func TestStringEncoding(t *testing.T) {
	cases := []struct {
		src string
		enc TokenEncoding
	}{
		{`"hello"`, EncodingASCII},
		{`u"hello"`, EncodingUnicode},
		{`u"café"`, EncodingUnicode},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			tok := lexOne(t, tc.src)
			if tok.Type != TokenString {
				t.Fatalf("Type: got %s, want %s", tok.Type, TokenString)
			}
			if tok.Encoding != tc.enc {
				t.Errorf("Encoding: got %q, want %q", tok.Encoding, tc.enc)
			}
		})
	}
}

func TestUnicodeStringContent(t *testing.T) {
	tok := lexOne(t, `u"café"`)
	if tok.Type != TokenString {
		t.Fatalf("Type: got %s, want %s", tok.Type, TokenString)
	}
	if tok.Encoding != EncodingUnicode {
		t.Errorf("Encoding: got %q, want %q", tok.Encoding, EncodingUnicode)
	}
	if tok.Text != `u"café"` {
		t.Errorf("Text: got %q, want %q", tok.Text, `u"café"`)
	}
	decoded, err := tok.Unquote()
	if err != nil {
		t.Fatalf("Unquote: unexpected error: %v", err)
	}
	if decoded != "café" {
		t.Errorf("Unquote: got %q, want %q", decoded, "café")
	}
}

func TestStringErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"unterminated_eof", `"hello`},
		{"unterminated_newline", "\"hello\nworld\""},
		{"unknown_escape_n", `"\n"`},
		{"unknown_escape_t", `"\t"`},
		{"unknown_escape_r", `"\r"`},
		{"unterminated_escape", "\"\\"},
		{"non_ascii", `"héllo"`},
		{"null_byte", "\"hello\x00world\""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Lex(tc.src)
			if err == nil {
				t.Errorf("Lex(%q): expected error, got nil", tc.src)
			}
		})
	}
}

func TestUnicodeStringErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"unterminated_eof", `u"hello`},
		{"unterminated_newline", "u\"hello\n\""},
		{"null_byte", "u\"hello\x00\""},
		{"unknown_escape", `u"\n"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Lex(tc.src)
			if err == nil {
				t.Errorf("Lex(%q): expected error, got nil", tc.src)
			}
		})
	}
}

func TestGlob(t *testing.T) {
	cases := []struct {
		src  string
		text string
	}{
		{"/", "/"},
		{"/usr/bin/ping", "/usr/bin/ping"},
		{"/usr/**/*.so", "/usr/**/*.so"},
		{"/proc/*/maps", "/proc/*/maps"},
		{"/dev/sd-a", "/dev/sd-a"},
		{"/var/log/app.log", "/var/log/app.log"},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			tok := lexOne(t, tc.src)
			if tok.Type != TokenString {
				t.Fatalf("Type: got %s, want %s", tok.Type, TokenString)
			}
			if tok.Encoding != EncodingASCII {
				t.Errorf("Encoding: got %q, want %q", tok.Encoding, EncodingASCII)
			}
			if tok.Text != tc.text {
				t.Errorf("Text: got %q, want %q", tok.Text, tc.text)
			}
		})
	}
}

func TestGlobTermination(t *testing.T) {
	cases := []struct {
		src      string
		globText string
		nextType TokenType
	}{
		{"/bin/sh =", "/bin/sh", TokenEq},
		{"/bin/sh,", "/bin/sh", TokenComma},
		{"/bin/sh(", "/bin/sh", TokenLParen},
		{"/bin/sh)", "/bin/sh", TokenRParen},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			toks := lexTokens(t, tc.src)
			if len(toks) < 2 {
				t.Fatalf("expected at least 2 tokens, got %d", len(toks))
			}
			if toks[0].Text != tc.globText {
				t.Errorf("glob Text: got %q, want %q", toks[0].Text, tc.globText)
			}
			if toks[1].Type != tc.nextType {
				t.Errorf("next token Type: got %s, want %s", toks[1].Type, tc.nextType)
			}
		})
	}
}

func TestGlobErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"question_mark", "/usr/bin/py?hon"},
		{"bracket_open", "/dev/sd[a-z]"},
		{"non_ascii", "/usr/héllo"},
		{"hash", "/usr/#private"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Lex(tc.src)
			if err == nil {
				t.Errorf("Lex(%q): expected error, got nil", tc.src)
			}
		})
	}
}

func TestTokenSequence(t *testing.T) {
	src := ".net read /proc/** where uid = 0"
	toks, err := Lex(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []struct {
		typ  TokenType
		text string
	}{
		{TokenSubsystem, "net"},
		{TokenName, "read"},
		{TokenString, "/proc/**"},
		{TokenWhere, "where"},
		{TokenName, "uid"},
		{TokenEq, "="},
		{TokenInt, "0"},
		{TokenEOF, ""},
	}
	if len(toks) != len(want) {
		t.Fatalf("token count: got %d, want %d: %v", len(toks), len(want), toks)
	}
	for i, w := range want {
		if toks[i].Type != w.typ {
			t.Errorf("toks[%d].Type: got %s, want %s", i, toks[i].Type, w.typ)
		}
		if toks[i].Text != w.text {
			t.Errorf("toks[%d].Text: got %q, want %q", i, toks[i].Text, w.text)
		}
	}
}

func TestAdjacentTokensNoWhitespace(t *testing.T) {
	// Operators and punctuation must tokenise without spaces between them.
	toks := lexTokens(t, "a>=b")
	if len(toks) != 3 {
		t.Fatalf("expected 3 tokens, got %d: %v", len(toks), toks)
	}
	want := []TokenType{TokenName, TokenGte, TokenName}
	for i, typ := range want {
		if toks[i].Type != typ {
			t.Errorf("toks[%d].Type: got %s, want %s", i, toks[i].Type, typ)
		}
	}
}

func TestUnquoteNonString(t *testing.T) {
	tok := lexOne(t, "foo")
	got, err := tok.Unquote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "foo" {
		t.Errorf("got %q, want %q", got, "foo")
	}
}

func TestUnquoteIntToken(t *testing.T) {
	tok := lexOne(t, "0x1F")
	got, err := tok.Unquote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "0x1F" {
		t.Errorf("got %q, want %q", got, "0x1F")
	}
}

func TestUnquoteGlob(t *testing.T) {
	tok := lexOne(t, "/usr/bin/*")
	got, err := tok.Unquote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/usr/bin/*" {
		t.Errorf("got %q, want %q", got, "/usr/bin/*")
	}
}

func TestUnquoteASCIIString(t *testing.T) {
	cases := []struct {
		src     string
		decoded string
	}{
		{`""`, ""},
		{`"hello"`, "hello"},
		{`"a\"b"`, `a"b`},
		{`"a\\b"`, `a\b`},
		{`"\\\""`, `\"`},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			tok := lexOne(t, tc.src)
			got, err := tok.Unquote()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.decoded {
				t.Errorf("got %q, want %q", got, tc.decoded)
			}
		})
	}
}

func TestUnquoteUnicodeString(t *testing.T) {
	tok := lexOne(t, `u"café\\"`)
	got, err := tok.Unquote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != `café\` {
		t.Errorf("got %q, want %q", got, `café\`)
	}
}

func TestUnquoteUnterminatedEscape(t *testing.T) {
	// The lexer never emits a token with a trailing backslash, so construct
	// one directly to exercise the error path in unescapeBody.
	tok := Token{Type: TokenString, Encoding: EncodingASCII, Text: `"ab\"`}
	_, err := tok.Unquote()
	if err == nil {
		t.Fatal("expected error for unterminated escape, got nil")
	}
}

func TestUnquoteUnknownEscape(t *testing.T) {
	// The lexer never emits a token with an unknown escape, so construct one
	// directly to exercise the error path in unescapeBody.
	tok := Token{Type: TokenString, Encoding: EncodingASCII, Text: `"a\nb"`}
	_, err := tok.Unquote()
	if err == nil {
		t.Fatal("expected error for unknown escape \\n, got nil")
	}
}

func TestQuantityLiteral(t *testing.T) {
	cases := []string{
		"1Ki", "1Mi", "1Gi", "1Ti", "1Pi", "1Ei",
		"1k", "1K", "1M", "1G", "1T", "1P", "1E",
		"500m", "100u", "5n",
		"0Gi", "1024Mi", "9999999999G",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			tok := lexOne(t, src)
			if tok.Type != TokenQuantity {
				t.Errorf("Type: got %s, want %s", tok.Type, TokenQuantity)
			}
			if tok.Text != src {
				t.Errorf("Text: got %q, want %q", tok.Text, src)
			}
		})
	}
}

func TestQuantityNotFused(t *testing.T) {
	// Suffix not in the recognised set: lexes as integer + name.
	cases := []struct {
		src      string
		wantTyps []TokenType
	}{
		{"1Gigabyte", []TokenType{TokenInt, TokenName}},
		{"500ms", []TokenType{TokenInt, TokenName}},
		{"1foo", []TokenType{TokenInt, TokenName}},
		{"0xff", []TokenType{TokenInt}}, // hex literal, no fusion
		{"0644", []TokenType{TokenInt}}, // legacy octal, no fusion
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			toks := lexTokens(t, tc.src)
			if len(toks) != len(tc.wantTyps) {
				t.Fatalf("token count: got %d, want %d (%v)", len(toks), len(tc.wantTyps), toks)
			}
			for i, want := range tc.wantTyps {
				if toks[i].Type != want {
					t.Errorf("token %d: got %s, want %s", i, toks[i].Type, want)
				}
			}
		})
	}
}

func TestQuantityFollowedByPunct(t *testing.T) {
	// A quantity that hits a clean terminator (whitespace, EOF, or grammar
	// punctuation) fuses; the trailing token is then lexed independently.
	toks := lexTokens(t, "1Gi 500m,2Mi")
	want := []TokenType{TokenQuantity, TokenQuantity, TokenComma, TokenQuantity}
	if len(toks) != len(want) {
		t.Fatalf("token count: got %d, want %d (%v)", len(toks), len(want), toks)
	}
	for i, w := range want {
		if toks[i].Type != w {
			t.Errorf("token %d: got %s, want %s", i, toks[i].Type, w)
		}
	}
}

func TestIsQuantitySuffix(t *testing.T) {
	for _, s := range []string{"Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "k", "K", "M", "G", "T", "P", "E", "m", "u", "n"} {
		if !isQuantitySuffix(s) {
			t.Errorf("isQuantitySuffix(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "x", "Xi", "kk", "Mib"} {
		if isQuantitySuffix(s) {
			t.Errorf("isQuantitySuffix(%q) = true, want false", s)
		}
	}
}

func TestHyphenInName(t *testing.T) {
	cases := []string{
		"eu-west-1",
		"api-stripe-com",
		"foo-bar",
		"a-b-c-d-e",
		"x-1",
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			tok := lexOne(t, src)
			if tok.Type != TokenName {
				t.Errorf("Type: got %s, want %s", tok.Type, TokenName)
			}
			if tok.Text != src {
				t.Errorf("Text: got %q, want %q", tok.Text, src)
			}
		})
	}
}

func TestHyphenBoundary(t *testing.T) {
	// A hyphen only joins when both sides are name characters. Trailing or
	// whitespace-flanked hyphens still tokenise as TokenMinus.
	cases := []struct {
		src      string
		wantTyps []TokenType
	}{
		{"foo-", []TokenType{TokenName, TokenMinus}},
		{"foo - bar", []TokenType{TokenName, TokenMinus, TokenName}},
		{"-foo", []TokenType{TokenMinus, TokenName}},
		{"foo-bar-", []TokenType{TokenName, TokenMinus}},
		{"a--b", []TokenType{TokenName, TokenMinus, TokenMinus, TokenName}},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			toks := lexTokens(t, tc.src)
			if len(toks) != len(tc.wantTyps) {
				t.Fatalf("token count: got %d, want %d (%v)", len(toks), len(tc.wantTyps), toks)
			}
			for i, w := range tc.wantTyps {
				if toks[i].Type != w {
					t.Errorf("token %d: got %s, want %s", i, toks[i].Type, w)
				}
			}
		})
	}
}

func TestHyphenInDottedName(t *testing.T) {
	// Hyphens combine freely with dots inside a single name token.
	tok := lexOne(t, "api.stripe-com.v1")
	if tok.Type != TokenName {
		t.Errorf("Type: got %s, want %s", tok.Type, TokenName)
	}
	if tok.Text != "api.stripe-com.v1" {
		t.Errorf("Text: got %q, want %q", tok.Text, "api.stripe-com.v1")
	}
}
