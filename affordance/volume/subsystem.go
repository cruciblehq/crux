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

// Returns the deduplication key for a volume grant.
//
// The key is the destination path. Two grants for the same destination are
// treated as a conflict because a path can only hold one volume mount.
func (s *Subsystem) Key(g *agl.Model) string {
	if len(g.Args) < 1 {
		return ""
	}
	return g.Args[0].Value
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
	readOnly := len(g.Args) < 2 || g.Args[1].Value == accessRead
	s.spec.Mounts = append(s.spec.Mounts, Mount{
		Destination: dest,
		ReadOnly:    readOnly,
	})
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
