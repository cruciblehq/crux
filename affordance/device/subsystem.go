package device

import (
	"os"
	"strconv"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
	"github.com/cruciblehq/crux/crex"
	"github.com/cruciblehq/crux/files"
)

// Implementation of the device node subsystem.
//
// Holds a pointer to the OCI device slice wired in at construction time. Each
// Build call appends one device node. This subsystem only provisions the node;
// access to it is gated separately by the cgroup device controller through the
// .cgroup subsystem.
type Subsystem struct {
	devices *[]specs.LinuxDevice
}

// Returns a Subsystem wired to append into devices.
func New(spec *[]specs.LinuxDevice) *Subsystem {
	return &Subsystem{devices: spec}
}

// Returns the device subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameDevice
}

// Returns the deduplication key for a device grant.
//
// The key is the device node path. Two grants for the same path are treated
// as a conflict because a path can only hold one device node.
func (s *Subsystem) Key(g *agl.Model) string {
	if len(g.Args) <= argPath {
		return ""
	}
	return g.Args[argPath].Value
}

// Applies a parsed grant to the wired-in devices slice.
//
// The grant has the form ".device TYPE PATH MAJOR MINOR [mode=OCTAL]
// [uid=UID] [gid=GID]". TYPE is one of c, b, u, or p. PATH is the absolute
// node path inside the container. MAJOR and MINOR are the device numbers.
// mode sets the node permission bits in octal; uid and gid set its owner.
// Access to the node is gated separately by .cgroup.
func (s *Subsystem) Build(g *agl.Model) error {
	if err := check(g); err != nil {
		return err
	}
	dev, err := parse(g)
	if err != nil {
		return err
	}
	*s.devices = append(*s.devices, dev)
	return nil
}

// Validates the structural shape of a device grant.
func check(g *agl.Model) error {
	if g.Where != nil {
		return crex.Wrapf(ErrInvalidGrant, "unexpected where clause in device grant")
	}
	if len(g.Args) != deviceArgCount {
		return crex.Wrapf(ErrInvalidGrant, "device grant requires exactly four arguments: type, path, major, and minor")
	}
	return nil
}

// Extracts and validates the grant's content into an OCI device node.
//
// Validates that the type is a known device type, the path is absolute, the
// major and minor numbers are integers, and any mode, uid, or gid keyword
// arguments are valid. mode is interpreted as octal.
func parse(g *agl.Model) (specs.LinuxDevice, error) {
	typeArg := g.Args[argType]
	if typeArg.Type != agl.ArgName {
		return specs.LinuxDevice{}, crex.Wrapf(ErrInvalidGrant, "first argument must be a device type name")
	}
	if _, ok := knownTypes[typeArg.Value]; !ok {
		return specs.LinuxDevice{}, crex.Wrapf(ErrInvalidGrant, "unknown device type %q; accepted types: c, b, u, p", typeArg.Value)
	}
	pathArg := g.Args[argPath]
	if pathArg.Type != agl.ArgName && pathArg.Type != agl.ArgStrASCII {
		return specs.LinuxDevice{}, crex.Wrapf(ErrInvalidGrant, "second argument must be a device node path")
	}
	nodePath, err := files.ValidateAbsPath(pathArg.Value)
	if err != nil {
		return specs.LinuxDevice{}, crex.Wrap(ErrInvalidGrant, err)
	}
	major, err := parseNumber(g.Args[argMajor], "major")
	if err != nil {
		return specs.LinuxDevice{}, err
	}
	minor, err := parseNumber(g.Args[argMinor], "minor")
	if err != nil {
		return specs.LinuxDevice{}, err
	}

	dev := specs.LinuxDevice{
		Path:  nodePath,
		Type:  typeArg.Value,
		Major: major,
		Minor: minor,
	}

	for _, kw := range g.Kwargs {
		if err := applyKwarg(&dev, kw); err != nil {
			return specs.LinuxDevice{}, err
		}
	}
	return dev, nil
}

// Applies a single keyword argument to the device node under construction.
func applyKwarg(dev *specs.LinuxDevice, kw agl.Kwarg) error {
	switch kw.Key {
	case kwMode:
		mode, err := parseMode(kw.Value)
		if err != nil {
			return err
		}
		dev.FileMode = &mode
	case kwUID:
		uid, err := parseID(kw.Value, kwUID)
		if err != nil {
			return err
		}
		dev.UID = &uid
	case kwGID:
		gid, err := parseID(kw.Value, kwGID)
		if err != nil {
			return err
		}
		dev.GID = &gid
	default:
		return crex.Wrapf(ErrInvalidGrant, "unknown keyword argument %q for device grant", kw.Key)
	}
	return nil
}

// Parses a device major or minor number from an integer argument.
func parseNumber(a agl.Arg, label string) (int64, error) {
	if a.Type != agl.ArgInt {
		return 0, crex.Wrapf(ErrInvalidGrant, "%s number must be an integer", label)
	}
	v, err := strconv.ParseInt(a.Value, decimalBase, deviceNumBits)
	if err != nil {
		return 0, crex.Wrapf(ErrInvalidGrant, "invalid %s number %q", label, a.Value)
	}
	return v, nil
}

// Parses an octal permission mode from a keyword-argument value.
func parseMode(a agl.Arg) (os.FileMode, error) {
	if a.Type != agl.ArgInt {
		return 0, crex.Wrapf(ErrInvalidGrant, "mode must be an octal integer")
	}
	v, err := strconv.ParseUint(a.Value, octalBase, idBits)
	if err != nil {
		return 0, crex.Wrapf(ErrInvalidGrant, "invalid octal mode %q", a.Value)
	}
	return os.FileMode(v), nil
}

// Parses an unsigned owner identifier from a keyword-argument value.
func parseID(a agl.Arg, label string) (uint32, error) {
	if a.Type != agl.ArgInt {
		return 0, crex.Wrapf(ErrInvalidGrant, "%s must be an integer", label)
	}
	v, err := strconv.ParseUint(a.Value, decimalBase, idBits)
	if err != nil {
		return 0, crex.Wrapf(ErrInvalidGrant, "invalid %s %q", label, a.Value)
	}
	return uint32(v), nil
}
