package aegis

import "errors"

var (
	ErrParse = errors.New("aegis parse failed")
	ErrLex   = errors.New("aegis lex failed")
)
