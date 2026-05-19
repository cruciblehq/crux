package fcap

import (
	pathpkg "path"
	"strings"

	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/manifest"

	"github.com/cruciblehq/crux/resource/affordance/agl"
	"github.com/cruciblehq/crux/resource/affordance/caps"
	"github.com/cruciblehq/crux/resource/affordance/subsystem"
)

// Implementation of the file capabilities subsystem.
//
// Wraps an fcap spec and routes incoming grants and merges into it. Each
// instance is bound to one spec for its lifetime; callers that need a new
// accumulator allocate a new spec and a new Subsystem.
type Subsystem struct {
	spec     *manifest.FcapSpec
	declared map[fcapKey]struct{} // Set of declared (cap, mode, path) triples.
}

// Identifies a single fcap declaration.
type fcapKey struct {
	cap  string
	mode manifest.FcapMode
	path string
}

// Returns a Subsystem bound to spec.
//
// Subsequent Build and Merge calls mutate spec in place. The caller retains
// ownership of spec and is responsible for any concurrency control; the
// Subsystem performs none.
func New(spec *manifest.FcapSpec) *Subsystem {
	return &Subsystem{spec: spec, declared: make(map[fcapKey]struct{})}
}

// Returns the fcap subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameFcap
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
	if err := s.declare(fcapKey{cap: cap, mode: mode, path: path}); err != nil {
		return err
	}
	return apply(s.spec, cap, mode, path)
}

// Records k as declared, returning ErrConflict if it was already declared.
func (s *Subsystem) declare(k fcapKey) error {
	if _, ok := s.declared[k]; ok {
		return crex.Wrapf(ErrConflict, "fcap %q %s %q already declared", k.cap, k.mode, k.path)
	}
	s.declared[k] = struct{}{}
	return nil
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
// CAP_ prefix), the second is a name naming a [manifest.FcapMode], and the third
// is a string or name carrying an absolute, NUL-free, already-clean path. The
// path is normalized through pathpkg.Clean and returned in its canonical form.
// All failures are wrapped as ErrInvalidGrant.
func parse(g *agl.Model) (string, manifest.FcapMode, string, error) {
	capArg := g.Args[0]
	if capArg.Type != agl.ArgName {
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "expected name as capability in fcap expression")
	}
	if _, err := caps.Parse(capArg.Value); err != nil {
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "unknown capability %q in fcap expression", capArg.Value)
	}
	modeArg := g.Args[1]
	if modeArg.Type != agl.ArgName {
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "expected name as mode in fcap expression")
	}
	mode := manifest.FcapMode(modeArg.Value)
	if !mode.IsValid() {
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "unknown fcap mode %q", modeArg.Value)
	}
	pathArg := g.Args[2]
	switch pathArg.Type {
	case agl.ArgStrASCII, agl.ArgStrUnicode, agl.ArgName:
	default:
		return "", "", "", crex.Wrapf(ErrInvalidGrant, "expected string as path in fcap expression")
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
func apply(s *manifest.FcapSpec, cap string, mode manifest.FcapMode, path string) error {
	if s.Entries == nil {
		s.Entries = make(map[string]*manifest.FcapCapabilities)
	}
	existing, ok := s.Entries[path]
	if !ok {
		existing = &manifest.FcapCapabilities{}
		s.Entries[path] = existing
	}
	switch mode {
	case manifest.FcapModeEffective:
		existing.GrantEffective([]string{cap})
	case manifest.FcapModeInheritable:
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
