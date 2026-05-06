package aegis

import "fmt"

// Canonical string form of unknown value types.
const unknownValueType = "<unknown>"

// Discriminates the type stored in a Value.
type ValueType int

const (
	ValueNone ValueType = iota // No value; uninitialised sentinel.
	ValueInt                   // Unsigned integer literal (decimal or 0x hex).
	ValueStr                   // String literal with escape sequences resolved.
	ValueVar                   // Variable reference; Str holds the name without the leading '$'.
)

// Source-text encoding of a string literal.
//
// ASCII strings are written as quoted "..." literals or as unquoted globs.
// Unicode strings are written with the u"..." prefix. The encoding is
// preserved on the AST so that consumers can distinguish between the two
// without re-parsing.
type StrEncoding int

const (
	StrASCII   StrEncoding = iota // Quoted ASCII literal or unquoted glob.
	StrUnicode                    // Quoted Unicode literal (u"..." prefix).
)

// Typed literal value carried by a non-field Operand.
//
// Either Int or Str is populated, depending on Type. When Type is ValueNone
// neither is populated and the Operand should not be inspected. StrEncoding
// is meaningful only when Type is ValueStr.
type Value struct {
	Type        ValueType   // Value type indicating which field holds the data.
	Int         uint64      // Unsigned 64-bit integer (valid when Type is ValueInt).
	Str         string      // Resolved string content (valid when Type is ValueStr).
	StrEncoding StrEncoding // Source-text encoding of the string literal (valid when Type is ValueStr).
}

// Renders the value in canonical source form.
//
// Integers are emitted in decimal regardless of their original base; strings
// are re-quoted with the standard escape rules and prefixed with 'u' when the
// encoding is Unicode; variables are rendered with the leading '$'. When Type
// is ValueNone the output is the canonical unknown value string.
func (v Value) String() string {
	switch v.Type {
	case ValueInt:
		return fmt.Sprintf("%d", v.Int)
	case ValueStr:
		if v.StrEncoding == StrUnicode {
			return fmt.Sprintf("u%q", v.Str)
		}
		return fmt.Sprintf("%q", v.Str)
	case ValueVar:
		return fmt.Sprintf("$%s", v.Str)
	default:
		return unknownValueType
	}
}
