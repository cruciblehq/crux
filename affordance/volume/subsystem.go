package volume

import (
	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/files"
)

// Implementation of the persistent storage volume subsystem.
//
// Holds a pointer to the accumulated [Spec] wired in at construction time.
// Each Build call appends one volume declaration; declarations are additive.
type Subsystem struct {
	spec *Spec // Write target for accumulated grants.
}

// Returns a Subsystem wired to accumulate into spec.
func New(spec *Spec) *Subsystem {
	return &Subsystem{spec: spec}
}

// Returns the volume subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameVolume
}

// Applies a parsed grant to the accumulated spec.
//
// The grant has the form ".volume DESTINATION [r|rw]". DESTINATION is an
// absolute path inside the container. The optional access token is "r" for a
// read-only mount or "rw" for a read-write mount; it defaults to "r" when
// omitted, keeping the baseline maximally restrictive. Any other value is
// rejected. No kwargs or where clause are accepted.
func (s *Subsystem) Build(g *agl.Model) error {
	if err := check(g); err != nil {
		return err
	}
	dest, err := files.ValidateAbsPath(g.Args[0].Value)
	if err != nil {
		return crex.Wrap(ErrInvalidGrant, err)
	}
	return s.upsertMount(buildMount(dest, g.Args))
}

// Builds a mount entry from parsed volume grant arguments.
func buildMount(dest string, args []agl.Arg) Mount {
	readOnly := len(args) < 2 || args[1].Value == accessRead
	return Mount{Destination: dest, ReadOnly: readOnly}
}

// Merges a mount declaration into the accumulated spec.
//
// Identical declarations are no-ops. A destination already declared with
// different options is rejected.
func (s *Subsystem) upsertMount(mount Mount) error {
	for _, existing := range s.spec.Mounts {
		if existing.Destination != mount.Destination {
			continue
		}
		if existing == mount {
			return nil
		}
		return crex.Newf(ErrInvalidGrant, "volume destination %q already declared with different options", mount.Destination)
	}
	s.spec.Mounts = append(s.spec.Mounts, mount)
	return nil
}

// Validates the structural shape of a volume grant.
func check(g *agl.Model) error {
	if g.Where != nil {
		return crex.Newf(ErrInvalidGrant, "unexpected where clause in volume grant")
	}
	if len(g.Kwargs) > 0 {
		return crex.Newf(ErrInvalidGrant, "unexpected keyword arguments in volume grant")
	}
	if len(g.Args) < 1 || len(g.Args) > 2 {
		return crex.Newf(ErrInvalidGrant, "volume grant requires one or two arguments")
	}
	if g.Args[0].Type != agl.ArgStrASCII && g.Args[0].Type != agl.ArgName {
		return crex.Newf(ErrInvalidGrant, "first argument must be a destination path")
	}
	if len(g.Args) == 2 {
		if g.Args[1].Type != agl.ArgName || (g.Args[1].Value != accessRead && g.Args[1].Value != accessReadWrite) {
			return crex.Newf(ErrInvalidGrant, "second argument must be %q (read-only) or %q (read-write)", "r", "rw")
		}
	}
	return nil
}
