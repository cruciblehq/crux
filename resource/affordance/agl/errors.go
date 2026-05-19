package agl

import "errors"

var (
	ErrParse = errors.New("agl parse failed")
	ErrLex   = errors.New("agl lex failed")
)
