package aegis

// Whether ch is a whitespace character.
//
// Only space and tab are recognised; newlines are not permitted since the
// grammar is single-line only.
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t'
}

// Whether ch is an ASCII decimal digit.
func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// Whether ch is an ASCII letter.
func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// Whether ch is an ASCII hexadecimal digit.
//
// Matches [0-9a-fA-F]; both lower and upper case letters are accepted.
func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// Whether ch is an ASCII octal digit.
//
// Matches [0-7].
func isOctalDigit(ch rune) bool {
	return ch >= '0' && ch <= '7'
}

// Whether ch is a binary digit.
//
// Matches 0 or 1 only.
func isBinaryDigit(ch rune) bool {
	return ch == '0' || ch == '1'
}

// Whether ch may begin a name.
//
// Matches [a-zA-Z_].
func isNameStart(ch rune) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// Whether ch may continue a name.
//
// Matches [a-zA-Z_0-9.]. Dots are permitted to support qualified field names
// such as "file.path".
func isNameContinue(ch rune) bool {
	return ch == '_' || ch == '.' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

// Whether ch may continue a variable name.
//
// Matches [a-zA-Z_0-9]. Dots are not permitted. The leading '$' is not part
// of the variable name and is not considered here since it is consumed
// separately.
func isVarContinue(ch rune) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

// Whether ch is a valid unquoted glob character.
//
// Matches [a-zA-Z0-9_./*-]. The glob metacharacter '*' is supported; '?' and
// '[...]' bracket expressions are not.
func isGlob(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_' || ch == '-' || ch == '.' || ch == '/' ||
		ch == '*'
}

// Whether ch cleanly terminates an unquoted glob without being part of it.
//
// Covers whitespace and the grammar punctuation and operator characters that
// can open a new token directly after a glob.
func isGlobTerminator(ch rune) bool {
	if isWhitespace(ch) {
		return true
	}
	switch ch {
	case ',', '(', ')', '=', '!', '>', '<', '&', '"':
		return true
	}
	return false
}
