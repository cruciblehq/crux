package agl

import "github.com/cruciblehq/crux/crex"

var (
	ErrParse = crex.New("agl parse failed")
	ErrLex   = crex.New("agl lex failed")
)
