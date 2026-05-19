package agl

import "testing"

func TestIsWhitespace(t *testing.T) {
	for _, ch := range " \t" {
		if !isWhitespace(ch) {
			t.Errorf("isWhitespace(%q) = false, want true", ch)
		}
	}
	for _, ch := range "\n\r\v\f" {
		if isWhitespace(ch) {
			t.Errorf("isWhitespace(%q) = true, want false", ch)
		}
	}
}

func TestIsDigit(t *testing.T) {
	for ch := rune('0'); ch <= '9'; ch++ {
		if !isDigit(ch) {
			t.Errorf("isDigit(%q) = false", ch)
		}
	}
	for _, ch := range "abcfAF/_x" {
		if isDigit(ch) {
			t.Errorf("isDigit(%q) = true, want false", ch)
		}
	}
}

func TestIsLetter(t *testing.T) {
	for ch := rune('a'); ch <= 'z'; ch++ {
		if !isLetter(ch) {
			t.Errorf("isLetter(%q) = false", ch)
		}
	}
	for ch := rune('A'); ch <= 'Z'; ch++ {
		if !isLetter(ch) {
			t.Errorf("isLetter(%q) = false", ch)
		}
	}
	for _, ch := range "0_-." {
		if isLetter(ch) {
			t.Errorf("isLetter(%q) = true, want false", ch)
		}
	}
}

func TestIsHexDigit(t *testing.T) {
	for _, ch := range "0123456789abcdefABCDEF" {
		if !isHexDigit(ch) {
			t.Errorf("isHexDigit(%q) = false", ch)
		}
	}
	for _, ch := range "ghijklmnopqrstuvwxyzGHIJKLMNOPQRSTUVWXYZ_" {
		if isHexDigit(ch) {
			t.Errorf("isHexDigit(%q) = true, want false", ch)
		}
	}
}

func TestIsOctalDigit(t *testing.T) {
	for ch := rune('0'); ch <= '7'; ch++ {
		if !isOctalDigit(ch) {
			t.Errorf("isOctalDigit(%q) = false", ch)
		}
	}
	for _, ch := range "89abcdef" {
		if isOctalDigit(ch) {
			t.Errorf("isOctalDigit(%q) = true, want false", ch)
		}
	}
}

func TestIsBinaryDigit(t *testing.T) {
	if !isBinaryDigit('0') {
		t.Error("isBinaryDigit('0') = false, want true")
	}
	if !isBinaryDigit('1') {
		t.Error("isBinaryDigit('1') = false, want true")
	}
	for _, ch := range "2345679abcABC_" {
		if isBinaryDigit(ch) {
			t.Errorf("isBinaryDigit(%q) = true, want false", ch)
		}
	}
}

func TestIsNameStart(t *testing.T) {
	for _, ch := range "_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		if !isNameStart(ch) {
			t.Errorf("isNameStart(%q) = false, want true", ch)
		}
	}
	for _, ch := range "0123456789.-@$" {
		if isNameStart(ch) {
			t.Errorf("isNameStart(%q) = true, want false", ch)
		}
	}
}

func TestIsNameContinue(t *testing.T) {
	for _, ch := range "_abcABC0123456789." {
		if !isNameContinue(ch) {
			t.Errorf("isNameContinue(%q) = false, want true", ch)
		}
	}
	for _, ch := range "-@$/" {
		if isNameContinue(ch) {
			t.Errorf("isNameContinue(%q) = true, want false", ch)
		}
	}
}

func TestIsVarContinue(t *testing.T) {
	for _, ch := range "_abcABC0123456789" {
		if !isVarContinue(ch) {
			t.Errorf("isVarContinue(%q) = false, want true", ch)
		}
	}
	// Dots are not permitted in variable names.
	for _, ch := range ".-@$/" {
		if isVarContinue(ch) {
			t.Errorf("isVarContinue(%q) = true, want false", ch)
		}
	}
}

func TestIsGlob(t *testing.T) {
	for _, ch := range "abcABC012_-./*" {
		if !isGlob(ch) {
			t.Errorf("isGlob(%q) = false, want true", ch)
		}
	}
	for _, ch := range " \t?[]()" {
		if isGlob(ch) {
			t.Errorf("isGlob(%q) = true, want false", ch)
		}
	}
}

func TestIsGlobTerminator(t *testing.T) {
	for _, ch := range " \t,()=!><&\"" {
		if !isGlobTerminator(ch) {
			t.Errorf("isGlobTerminator(%q) = false, want true", ch)
		}
	}
	for _, ch := range "abcABC012_-.*/" {
		if isGlobTerminator(ch) {
			t.Errorf("isGlobTerminator(%q) = true, want false", ch)
		}
	}
}
