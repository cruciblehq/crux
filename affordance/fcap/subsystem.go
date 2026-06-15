package fcap

import (
	"fmt"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/capset"
	"github.com/cruciblehq/crux/affordance/subsystem"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/files"
)

// Implementation of the file capabilities subsystem.
//
// Wraps an fcap spec and routes incoming grants and merges into it. Each
// instance is bound to one spec for its lifetime; callers that need a new
// accumulator allocate a new spec and a new Subsystem.
type Subsystem struct {
	spec *Spec
}

// Returns a Subsystem bound to spec.
//
// Subsequent Build and Merge calls mutate spec in place. The caller retains
// ownership of spec and is responsible for any concurrency control; the
// Subsystem performs none.
func New(spec *Spec) *Subsystem {
	return &Subsystem{spec: spec}
}

// Returns the fcap subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameFcap
}

// Returns the deduplication key for an fcap grant.
//
// The key is the capability name and absolute path joined by ":". Mode is
// intentionally excluded so that granting the same capability on the same
// file with a different mode (e.g. permitted vs inheritable) is still
// treated as a conflict.
func (s *Subsystem) Key(g *agl.Model) string {
	if len(g.Args) < 3 {
		return ""
	}
	return fmt.Sprintf("%s:%s", g.Args[0].Value, g.Args[2].Value)
}

// Applies a parsed grant to the wired-in section.
//
// The grant has the form ".fcap CAP MODE PATH" where CAP is a capability
// name without the CAP_ prefix, MODE is one of "effective" or "inheritable",
// and PATH is an absolute, clean filesystem path.
func (s *Subsystem) Build(g *agl.Model) error {
	if err := check(g); err != nil {
		return err
	}
	cap, mode, path, err := parse(g)
	if err != nil {
		return err
	}
	return apply(s.spec, cap, mode, path)
}

// Validates the grant's structural shape against what the fcap subsystem accepts.
//
// Rejects grants that carry a where clause, keyword arguments, or a
// positional arity other than three. Per-argument typing and value
// validation are deferred to parse.
func check(g *agl.Model) error {
	if g.Where != nil {
		return crex.Wrapf(ErrInvalidGrant, "unexpected where clause in fcap expression")
	}
	if len(g.Kwargs) != 0 {
		return crex.Wrapf(ErrInvalidGrant, "unexpected keyword arguments in fcap expression")
	}
	if len(g.Args) != 3 {
		return crex.Wrapf(ErrInvalidGrant, "wrong number of arguments in fcap expression")
	}
	return nil
}

// Extracts and validates the grant's capability, mode, and target path.
//
// The first argument is a name resolving to a known capability (without the
// CAP_ prefix), the second is a name naming a [FcapMode], and the third is a
// string or name carrying an absolute, NUL-free, already-clean path. The path
// is validated and returned in its canonical form by [files.ValidateAbsPath].
// All failures are wrapped as ErrInvalidGrant.
func parse(g *agl.Model) (string, Mode, string, error) {
	capArg := g.Args[0]
	if capArg.Type != agl.ArgName {
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "expected name as capability in fcap expression")
	}
	if _, err := capset.Parse(capArg.Value); err != nil {
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "unknown capability %q in fcap expression", capArg.Value)
	}
	modeArg := g.Args[1]
	if modeArg.Type != agl.ArgName {
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "expected name as mode in fcap expression")
	}
	mode, err := ParseMode(modeArg.Value)
	if err != nil {
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "unknown fcap mode %q", modeArg.Value)
	}
	pathArg := g.Args[2]
	switch pathArg.Type {
	case agl.ArgStrASCII, agl.ArgStrUnicode, agl.ArgName:
	default:
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "expected string as path in fcap expression")
	}
	path, err := files.ValidateAbsPath(pathArg.Value)
	if err != nil {
		return "", "", "", crex.Wrap(ErrInvalidGrant, err)
	}
	return capArg.Value, mode, path, nil
}

// Applies a parsed capability/mode/path triple to s.
//
// Creates the per-path Capabilities entry on first use and folds cap into it
// according to mode: ModeEffective adds to the file-permitted set and raises
// the effective bit so the capability is live immediately after execve;
// ModeInheritable adds to the file-inheritable set, which only takes effect
// if the exec caller also holds the capability in its own inheritable set.
// Repeated grants for the same (path, cap, mode) are idempotent.
func apply(s *Spec, cap string, mode Mode, path string) error {
	if s.Entries == nil {
		s.Entries = make(map[string]*Capabilities)
	}
	existing, ok := s.Entries[path]
	if !ok {
		existing = &Capabilities{}
		s.Entries[path] = existing
	}
	switch mode {
	case ModeEffective:
		existing.GrantEffective([]string{cap})
	case ModeInheritable:
		existing.GrantInheritable([]string{cap})
	}
	return nil
}
