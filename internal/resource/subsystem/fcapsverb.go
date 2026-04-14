package subsystem

import "github.com/cruciblehq/crex"

// File capability verb.
type fcapVerb string

const (
	fcapVerbEffective   fcapVerb = "effective"   // File-permitted + effective bit.
	fcapVerbInheritable fcapVerb = "inheritable" // File-inheritable only.
)

// Converts a string to an fcapVerb, returning an error for unknown values.
func parseFcapVerb(s string) (fcapVerb, error) {
	v := fcapVerb(s)
	switch v {
	case fcapVerbEffective, fcapVerbInheritable:
		return v, nil
	default:
		return "", crex.Wrapf(ErrGrantExpression, "unknown verb %q", s)
	}
}
