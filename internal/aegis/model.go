package aegis

import (
	"strings"
)

// Top-level AST node produced by [Parse] from a domain grant source string.
//
// Subsystem holds the name following the leading "." in the source. Args is
// the non-empty positional argument list that follows. Kwargs holds zero or
// more key=value pairs after the positionals. Where holds the optional
// where-clause predicate, nil when absent. The struct is the parsed view of
// a grant; it does not represent reference grants since those never reach
// the parser.
type Model struct {
	Subsystem string  // Subsystem name without the leading ".".
	Args      []Arg   // Positional arguments. Always non-empty after a successful Parse.
	Kwargs    []Kwarg // Keyword arguments.
	Where     Expr    // Where-clause filter predicate, or nil.
}

// Renders the canonical source form of the parsed grant.
//
// Output has the form ".subsystem args... kwargs... [where expr]" with
// whitespace normalised, integer literals in decimal, strings re-quoted with
// the standard escape rules, and where-clause parentheses reflecting AST
// structure rather than the original grouping. The result is parseable by
// [Parse] and yields an equivalent Model.
func (p *Model) String() string {
	var b strings.Builder
	b.WriteByte('.')
	b.WriteString(p.Subsystem)
	for _, a := range p.Args {
		b.WriteByte(' ')
		b.WriteString(a.String())
	}
	for _, k := range p.Kwargs {
		b.WriteByte(' ')
		b.WriteString(k.String())
	}
	if p.Where != nil {
		b.WriteString(" where ")
		b.WriteString(p.Where.String())
	}
	return b.String()
}
