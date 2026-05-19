package agl

// Map of reserved keywords to their token types.
//
// Consulted after lexing a name-shaped token to determine whether it is a
// reserved keyword or a plain name. Only lower-case forms are recognised;
// keywords are case-sensitive.
var keywords = map[string]TokenType{
	"where":   TokenWhere,
	"and":     TokenAnd,
	"or":      TokenOr,
	"not":     TokenNot,
	"in":      TokenIn,
	"like":    TokenLike,
	"between": TokenBetween,
}

// Whether s is a reserved keyword.
func isKeyword(s string) bool {
	_, ok := keywords[s]
	return ok
}
