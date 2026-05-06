package aegis

import "fmt"

// Canonical string form of unknown argument types.
const unknownArgType = "<unknown>"

// Lexical category of an Arg or Kwarg as it appears in the source.
//
// Preserved on the AST so consumers can tell symbol types apart (a name from an
// integer, ASCII from Unicode, etc) without re-parsing. Subsystems re-parse the
// Value text as needed to obtain a typed integer or to distinguish a glob from
// a quoted string.
type ArgType int

const (
	ArgName       ArgType = iota // Name token (identifier, dotted name).
	ArgInt                       // Unsigned integer literal in a C-style base.
	ArgQuantity                  // Unsigned integer literal with a unit suffix (e.g. 1Gi, 500m).
	ArgStrASCII                  // Quoted ASCII string literal or unquoted glob.
	ArgStrUnicode                // Quoted Unicode string literal (u"..." prefix).
	ArgVar                       // Variable reference (Value holds the name without the leading '$').
)

// Single positional argument or keyword-argument value.
//
// Type names the lexical category the argument was written as. Value holds the
// decoded text: for string types escape sequences (\" and \\) already resolved
// and surrounding quotes stripped; for names and integers it is the source text.
// Subsystems re-parse Value when they need a typed integer.
type Arg struct {
	Type  ArgType // Lexical category.
	Value string  // Decoded text content of the argument.
}

// Renders the argument in canonical source form.
//
// Strings are re-quoted with the standard escape rules; Unicode strings carry
// the u"..." prefix; variables are rendered with the leading '$'; names and
// integers are emitted verbatim.
func (a Arg) String() string {
	switch a.Type {
	case ArgName, ArgInt, ArgQuantity:
		return a.Value
	case ArgStrASCII:
		return fmt.Sprintf("%q", a.Value)
	case ArgStrUnicode:
		return fmt.Sprintf("u%q", a.Value)
	case ArgVar:
		return fmt.Sprintf("$%s", a.Value)
	default:
		return unknownArgType
	}
}

// A single key=value keyword argument.
//
// Key is a NAME (possibly dotted, e.g. "cpu.weight"). Value carries the type
// and decoded text of the RHS scalar. Kwargs only appear after all positional
// arguments within a grant.
type Kwarg struct {
	Key   string // NAME on the LHS.
	Value Arg    // RHS scalar.
}

// Renders the kwarg in canonical "key=value" form.
func (k Kwarg) String() string {
	return fmt.Sprintf("%s=%s", k.Key, k.Value)
}
