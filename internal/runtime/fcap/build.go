package fcap

import (
	pathpkg "path"
	"strings"

	"github.com/cruciblehq/crux/internal/crex"

	"github.com/cruciblehq/crux/internal/manifest/grant"
	"github.com/cruciblehq/crux/internal/runtime/shared"
	fcapspec "github.com/cruciblehq/crux/internal/runtime/shared/fcap"
)

// Implementation of the file capabilities subsystem.
//
// Wraps an fcap spec and routes incoming grants and merges into it. Each
// instance is bound to one spec for its lifetime; callers that need a
// fresh accumulator allocate a new spec and a new Subsystem.
type Subsystem struct {
	spec *fcapspec.Spec
}

// Returns a Subsystem bound to spec.
//
// Subsequent Build and Merge calls mutate spec in place. The caller retains
// ownership of spec and is responsible for any concurrency control; the
// Subsystem performs none.
func New(spec *fcapspec.Spec) *Subsystem {
	return &Subsystem{spec: spec}
}

// Returns the fcap subsystem identifier.
func (s *Subsystem) Name() shared.Name {
	return shared.NameFcap
}

// Applies a parsed grant to the wired-in section.
//
// The grant has the form ".fcap CAP MODE PATH" where CAP is a capability
// name without the CAP_ prefix, MODE is one of "effective" or "inheritable",
// and PATH is an absolute, clean filesystem path.
func (s *Subsystem) Build(g grant.Grant) error {
	if err := check(&g); err != nil {
		return err
	}
	cap, mode, path, err := parse(&g)
	if err != nil {
		return err
	}
	return apply(s.spec, cap, mode, path)
}

// Folds the fcap section of src into the bound spec.
//
// Per-path capability lists are unioned and the effective bit is OR'd, so
// merging is idempotent and order-independent. A nil src.Fcap is a no-op.
func (s *Subsystem) Merge(src shared.Spec) error {
	if src.Fcap == nil {
		return nil
	}
	s.spec.MergeSpec(src.Fcap)
	return nil
}

// Validates the grant's structural shape against what the fcap subsystem accepts.
//
// Rejects grants that carry a where clause, keyword arguments, or a
// positional arity other than three. Per-argument typing and value
// validation are deferred to parse.
func check(g *grant.Grant) error {
	if g.Where != nil {
		return crex.Wrapf(ErrInvalidGrant, "unexpected where clause in fcap expression")
	}
	if len(g.Kwargs) != 0 {
		return crex.Wrapf(ErrInvalidGrant, "unexpected keyword arguments in fcap expression")
	}
	if len(g.Args) != 3 {
		return crex.Wrapf(ErrInvalidGrant, "wrong number of arguments in fcap expression: got %d, want cap mode path", len(g.Args))
	}
	return nil
}

// Extracts and validates the grant's capability, mode, and target path.
//
// The first argument is a name resolving to a known capability (without the
// CAP_ prefix), the second is a name naming a [fcapspec.Mode], and the third
// is a string or name carrying an absolute, NUL-free, already-clean path. The
// path is normalized through pathpkg.Clean and returned in its canonical form.
// All failures are wrapped as ErrInvalidGrant.
func parse(g *grant.Grant) (string, fcapspec.Mode, string, error) {
	capArg := g.Args[0]
	if capArg.Type != grant.ArgName {
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "expected name as capability in fcap expression, got %s", capArg)
	}
	if _, err := shared.ParseCap(capArg.Value); err != nil {
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "unknown capability %q in fcap expression", capArg.Value)
	}
	modeArg := g.Args[1]
	if modeArg.Type != grant.ArgName {
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "expected name as mode in fcap expression, got %s", modeArg)
	}
	mode := fcapspec.Mode(modeArg.Value)
	if !mode.IsValid() {
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "unknown fcap mode %q", modeArg.Value)
	}
	pathArg := g.Args[2]
	switch pathArg.Type {
	case grant.ArgStrASCII, grant.ArgStrUnicode, grant.ArgName:
	default:
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "expected string as path in fcap expression, got %s", pathArg)
	}
	path, err := normalizePath(pathArg.Value)
	if err != nil {
		return "", "", "", err
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
func apply(s *fcapspec.Spec, cap string, mode fcapspec.Mode, path string) error {
	if s.Entries == nil {
		s.Entries = make(map[string]*fcapspec.Capabilities)
	}
	existing, ok := s.Entries[path]
	if !ok {
		existing = &fcapspec.Capabilities{}
		s.Entries[path] = existing
	}
	switch mode {
	case fcapspec.ModeEffective:
		existing.GrantEffective([]string{cap})
	case fcapspec.ModeInheritable:
		existing.GrantInheritable([]string{cap})
	}
	return nil
}

// Validates a binary path.
//
// The path must be non-empty, absolute, not have a trailing slash, contain
// no NUL bytes, and already be clean.
func normalizePath(path string) (string, error) {
	if path == "" {
		return "", crex.Wrapf(ErrInvalidGrant, "path is empty")
	}
	if strings.Contains(path, "\x00") {
		return "", crex.Wrapf(ErrInvalidGrant, "path %q contains NUL", path)
	}
	if !pathpkg.IsAbs(path) {
		return "", crex.Wrapf(ErrInvalidGrant, "path %q must be absolute", path)
	}
	if strings.HasSuffix(path, "/") {
		return "", crex.Wrapf(ErrInvalidGrant, "path %q must not have a trailing slash", path)
	}
	clean := pathpkg.Clean(path)
	if path != clean {
		return "", crex.Wrapf(ErrInvalidGrant, "path %q must be clean", path)
	}
	return clean, nil
}
