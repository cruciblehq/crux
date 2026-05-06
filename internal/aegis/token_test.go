package aegis

import (
	"errors"
	"testing"
)

func TestTokenTypeIsOperator(t *testing.T) {
	operators := []TokenType{
		TokenEq, TokenNeq,
		TokenGt, TokenGte,
		TokenLt, TokenLte,
		TokenAmpersand,
	}
	for _, tt := range operators {
		if !tt.isOperator() {
			t.Errorf("%s.isOperator() = false, want true", tt)
		}
	}
	nonOperators := []TokenType{
		TokenSubsystem, TokenName, TokenString,
		TokenInt, TokenQuantity, TokenVar,
		TokenWhere, TokenAnd, TokenOr, TokenNot,
		TokenIn, TokenLike, TokenBetween,
		TokenMinus, TokenLParen, TokenRParen, TokenComma,
		TokenError, TokenEOF,
	}
	for _, tt := range nonOperators {
		if tt.isOperator() {
			t.Errorf("%s.isOperator() = true, want false", tt)
		}
	}
}

func TestTokenUnquoteNonString(t *testing.T) {
	// Non-string tokens return Text unchanged regardless of content.
	for _, tt := range []TokenType{TokenName, TokenInt, TokenEOF, TokenQuantity} {
		tok := Token{Type: tt, Text: "hello"}
		got, err := tok.Unquote()
		if err != nil {
			t.Errorf("%s.Unquote(): unexpected error: %v", tt, err)
		}
		if got != "hello" {
			t.Errorf("%s.Unquote() = %q, want %q", tt, got, "hello")
		}
	}
}

func TestTokenUnquoteGlob(t *testing.T) {
	// Unquoted glob strings (starting with '/') are returned verbatim.
	tok := Token{Type: TokenString, Text: "/usr/bin/**"}
	got, err := tok.Unquote()
	if err != nil {
		t.Fatalf("Unquote(): unexpected error: %v", err)
	}
	if got != "/usr/bin/**" {
		t.Errorf("got %q, want %q", got, "/usr/bin/**")
	}
}

func TestTokenUnquoteASCII(t *testing.T) {
	cases := []struct {
		text    string
		decoded string
	}{
		{`""`, ""},
		{`"hello"`, "hello"},
		{`"say \"hi\""`, `say "hi"`},
		{`"a\\b"`, `a\b`},
	}
	for _, tc := range cases {
		tok := Token{Type: TokenString, Encoding: EncodingASCII, Text: tc.text}
		got, err := tok.Unquote()
		if err != nil {
			t.Errorf("Unquote(%q): unexpected error: %v", tc.text, err)
			continue
		}
		if got != tc.decoded {
			t.Errorf("Unquote(%q) = %q, want %q", tc.text, got, tc.decoded)
		}
	}
}

func TestTokenUnquoteUnicode(t *testing.T) {
	tok := Token{Type: TokenString, Encoding: EncodingUnicode, Text: `u"café"`}
	got, err := tok.Unquote()
	if err != nil {
		t.Fatalf("Unquote(): unexpected error: %v", err)
	}
	if got != "café" {
		t.Errorf("got %q, want %q", got, "café")
	}
}

func TestTokenUnquoteErrors(t *testing.T) {
	cases := []struct {
		name string
		tok  Token
	}{
		{
			"empty",
			Token{Type: TokenString, Text: ""},
		},
		{
			"no_closing_quote",
			Token{Type: TokenString, Encoding: EncodingASCII, Text: `"hello`},
		},
		{
			"unknown_escape",
			Token{Type: TokenString, Encoding: EncodingASCII, Text: `"\n"`},
		},
		{
			"unterminated_escape",
			Token{Type: TokenString, Encoding: EncodingASCII, Text: `"\`},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.tok.Unquote()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, ErrLex) {
				t.Errorf("error %v does not wrap ErrLex", err)
			}
		})
	}
}
