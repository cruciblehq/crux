package mount

import (
	"fmt"
	"strconv"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
	"github.com/cruciblehq/crux/affordance/units"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/files"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Implementation of the kernel VFS mount subsystem.
//
// Holds a pointer to the OCI Mounts slice wired in at construction time.
// Each Build call appends one mount entry.
type Subsystem struct {
	mounts *[]specs.Mount
}

// Returns a Subsystem wired to append into mounts.
func New(spec *[]specs.Mount) *Subsystem {
	return &Subsystem{mounts: spec}
}

// Returns the mount subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameMount
}

// Returns the deduplication key for a mount grant.
//
// The key is the destination path. Two grants for the same destination —
// regardless of type — are treated as a conflict because a path can only
// hold one mount.
func (s *Subsystem) Key(g *agl.Model) string {
	if len(g.Args) <= argDest {
		return ""
	}
	return g.Args[argDest].Value
}

// Applies a parsed grant to the wired-in mounts slice.
//
// The grant has the form ".mount TYPE DESTINATION [size=QUANTITY] [mode=OCTAL]".
// TYPE must be one of the accepted in-kernel filesystem types. DESTINATION is
// an absolute path inside the container. size and mode are optional keyword
// arguments valid only for tmpfs; size sets an upper bound on the filesystem's
// memory usage and mode sets the root directory permission bits.
func (s *Subsystem) Build(g *agl.Model) error {
	if err := check(g); err != nil {
		return err
	}
	fsType := g.Args[argType].Value
	dest, err := files.ValidateAbsPath(g.Args[argDest].Value)
	if err != nil {
		return crex.Wrap(ErrInvalidGrant, err)
	}

	opts, err := buildOptions(fsType, g.Kwargs)
	if err != nil {
		return err
	}

	*s.mounts = append(*s.mounts, specs.Mount{
		Type:        fsType,
		Source:      fsType,
		Destination: dest,
		Options:     opts,
	})
	return nil
}

// Validates the structural shape of a mount grant.
func check(g *agl.Model) error {
	if g.Where != nil {
		return crex.Newf(ErrInvalidGrant, "unexpected where clause in mount grant")
	}
	if len(g.Args) != mountArgCount {
		return crex.Newf(ErrInvalidGrant, "mount grant requires exactly two arguments: type and destination")
	}
	if g.Args[argType].Type != agl.ArgName {
		return crex.Newf(ErrInvalidGrant, "first argument must be a filesystem type name")
	}
	fsType := g.Args[argType].Value
	if _, ok := knownTypes[fsType]; !ok {
		return crex.Newf(ErrInvalidGrant, "unknown filesystem type %q (expected tmpfs, proc, sysfs, devpts, mqueue, or cgroup2)", fsType)
	}
	if g.Args[argDest].Type != agl.ArgName && g.Args[argDest].Type != agl.ArgStrASCII {
		return crex.Newf(ErrInvalidGrant, "second argument must be a destination path")
	}
	return nil
}

// Builds the mount options list for the given filesystem type and kwargs.
//
// All mounts receive the nosuid, nodev, noexec baseline. tmpfs mounts also
// receive a default mode=1777; the mode kwarg overrides it and the size kwarg
// adds a size limit. No kwargs are accepted for non-tmpfs types.
func buildOptions(fsType string, kwargs []agl.Kwarg) ([]string, error) {
	base := []string{optNosuid, optNodev, optNoexec}
	if fsType != fsTmpfs {
		if len(kwargs) > 0 {
			return nil, crex.Newf(ErrInvalidGrant, "keyword arguments are not supported for %s mounts", fsType)
		}
		return base, nil
	}

	// tmpfs: start with a default mode that the mode= kwarg can override.
	mode := defaultTmpfsMode
	var extras []string
	for _, kw := range kwargs {
		switch kw.Key {
		case kwSize:
			bytes, err := parseSize(kw.Value)
			if err != nil {
				return nil, err
			}
			extras = append(extras, fmt.Sprintf("%s=%d", kwSize, bytes))
		case kwMode:
			m, err := parseMode(kw.Value)
			if err != nil {
				return nil, err
			}
			mode = m
		default:
			return nil, crex.Newf(ErrInvalidGrant, "unknown keyword argument %q for tmpfs mount (expected size or mode)", kw.Key)
		}
	}

	opts := append(base, fmt.Sprintf("%s=%s", kwMode, mode)) //nolint:gocritic // intentional append
	return append(opts, extras...), nil
}

// Validates a tmpfs mode keyword value and returns it in canonical form.
//
// The value must be an integer interpreted as octal, mirroring the device
// subsystem's mode validation. Returns the original octal spelling so that it
// can be emitted verbatim into the mount options.
func parseMode(a agl.Arg) (string, error) {
	if a.Type != agl.ArgInt {
		return "", crex.Newf(ErrInvalidGrant, "mode must be an octal integer")
	}
	if _, err := strconv.ParseUint(a.Value, 8, 32); err != nil {
		return "", crex.Newf(ErrInvalidGrant, "invalid octal mode %q", a.Value)
	}
	return a.Value, nil
}

// Parses a tmpfs size keyword value into a byte count.
//
// Accepts an integer literal as raw bytes or a quantity with an integer
// multiplier suffix (Ki-Ei binary, k/K-E decimal). Sub-unit suffixes (m, u, n)
// have no integer multiplier and are rejected.
func parseSize(a agl.Arg) (uint64, error) {
	switch a.Type {
	case agl.ArgInt:
		n, err := strconv.ParseUint(a.Value, 10, 64)
		if err != nil {
			return 0, crex.Newf(ErrInvalidGrant, "invalid size %q", a.Value)
		}
		return n, nil
	case agl.ArgQuantity:
		for _, n := range []int{2, 1} {
			if len(a.Value) <= n {
				continue
			}
			suffix := units.QuantitySuffix(a.Value[len(a.Value)-n:])
			mul, ok := suffix.Multiplier()
			if !ok {
				continue
			}
			v, err := strconv.ParseUint(a.Value[:len(a.Value)-n], 10, 64)
			if err != nil {
				return 0, crex.Newf(ErrInvalidGrant, "invalid size %q", a.Value)
			}
			return v * mul, nil
		}
		return 0, crex.Newf(ErrInvalidGrant, "unknown size suffix in %q", a.Value)
	default:
		return 0, crex.Newf(ErrInvalidGrant, "size must be a byte quantity")
	}
}
