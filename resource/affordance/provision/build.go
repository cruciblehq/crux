package provision

import (
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/crex"

	"github.com/cruciblehq/crux/resource/affordance/agl"
	"github.com/cruciblehq/crux/resource/affordance/subsystem"
	"github.com/cruciblehq/crux/resource/affordance/units"
)

// Implementation of the provision subsystem.
//
// Declares CPU, memory, and disk allocation from .provision grants into a
// [Spec]. The blueprint builder reads the spec to bin-pack services onto
// compute units and select instance types.
type Subsystem struct {
	spec *Spec // Pointer to the unified spec's provision section.
}

// Returns a Subsystem that writes into s.
func New(s *Spec) *Subsystem {
	return &Subsystem{spec: s}
}

// Returns the provision subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameProvision
}

// Returns the deduplication key for a provision grant.
//
// The key is the resource name from args[0] (e.g. "cpu", "memory", "disk").
// Duplicate resource declarations are treated as a conflict; grants are not
// additive for the same resource type.
func (s *Subsystem) Key(g *agl.Model) string {
	if len(g.Args) == 0 {
		return ""
	}
	return g.Args[0].Value
}

// Applies a single .provision grant to the accumulated spec.
//
// Grants are of the form ".provision RESOURCE VALUE". RESOURCE is one of "cpu",
// "memory", or "disk". VALUE is a quantity appropriate for that resource. Each
// resource may appear at most once. Duplicate declarations of the same resource
// must be rejected before reaching this method. The builder's Spec is mutated
// in-place with the new provision values.
func (s *Subsystem) Build(g *agl.Model) error {
	if err := check(g); err != nil {
		return err
	}
	resource := g.Args[0].Value
	value := g.Args[1]
	switch resource {
	case ResourceCPU:
		v, err := parseCPU(value)
		if err != nil {
			return err
		}
		s.spec.CPU = v
	case ResourceMemory:
		v, err := parseBytes(value)
		if err != nil {
			return err
		}
		s.spec.Memory = v
	case ResourceDisk:
		v, err := parseBytes(value)
		if err != nil {
			return err
		}
		s.spec.Disk = v
	default:
		return crex.Wrapf(ErrInvalidGrant, "unknown resource %q in provision grant", resource)
	}
	return nil
}

// Validates the structural shape of a provision grant.
//
// Provision grants take exactly two positional arguments (resource name and
// value) and no keyword arguments or where clause.
func check(g *agl.Model) error {
	if g.Where != nil {
		return crex.Wrapf(ErrInvalidGrant, "unexpected where clause in provision grant")
	}
	if len(g.Kwargs) > 0 {
		return crex.Wrapf(ErrInvalidGrant, "unexpected keyword arguments in provision grant")
	}
	if len(g.Args) != 2 {
		return crex.Wrapf(ErrInvalidGrant, "provision grant requires exactly two arguments (resource name and value)")
	}
	if g.Args[0].Type != agl.ArgName {
		return crex.Wrapf(ErrInvalidGrant, "first argument must be a resource name (cpu, memory, disk)")
	}
	return nil
}

// Parses a CPU argument into millicores.
//
// An integer arg (e.g. "4") is converted to millicores. A quantity arg with
// the "m" suffix (e.g. "500m") is stored directly. Other types and suffixes
// are rejected.
func parseCPU(a agl.Arg) (uint64, error) {
	switch a.Type {
	case agl.ArgInt:
		n, err := strconv.ParseUint(a.Value, 10, 64)
		if err != nil {
			return 0, crex.Wrap(ErrInvalidGrant, err)
		}
		return n * 1000, nil
	case agl.ArgQuantity:
		if !strings.HasSuffix(a.Value, string(units.SuffixMilli)) {
			return 0, crex.Wrapf(ErrInvalidGrant, "cpu quantity must use the millicore suffix")
		}
		n, err := strconv.ParseUint(strings.TrimSuffix(a.Value, string(units.SuffixMilli)), 10, 64)
		if err != nil {
			return 0, crex.Wrap(ErrInvalidGrant, err)
		}
		return n, nil
	default:
		return 0, crex.Wrapf(ErrInvalidGrant, "cpu must be a vCPU count or millicore quantity")
	}
}

// Parses a byte quantity argument into bytes.
//
// Accepts an integer or quantity arg with IEC binary (Ki, Mi, Gi, Ti, Pi, Ei)
// or SI decimal (k/K, M, G, T, P) suffixes.
func parseBytes(a agl.Arg) (uint64, error) {
	switch a.Type {
	case agl.ArgInt:
		n, err := strconv.ParseUint(a.Value, 10, 64)
		if err != nil {
			return 0, crex.Wrap(ErrInvalidGrant, err)
		}
		return n, nil
	case agl.ArgQuantity:
		return parseByteQuantity(a.Value)
	default:
		return 0, crex.Wrapf(ErrInvalidGrant, "expected a byte quantity")
	}
}

// Converts a quantity string (e.g. "8Gi") to bytes.
//
// Tries the two-character IEC binary suffixes (Ki–Ei) before the one-character
// SI decimal suffixes (k/K–P). Sub-unit suffixes (m, u, n) are rejected since
// they cannot be expressed as whole bytes.
func parseByteQuantity(s string) (uint64, error) {
	for _, n := range []int{2, 1} {
		if len(s) <= n {
			continue
		}
		suffix := units.QuantitySuffix(s[len(s)-n:])
		mul, ok := suffix.Multiplier()
		if !ok {
			continue
		}
		v, err := strconv.ParseUint(s[:len(s)-n], 10, 64)
		if err != nil {
			return 0, crex.Wrap(ErrInvalidGrant, err)
		}
		return v * mul, nil
	}
	return 0, crex.Wrapf(ErrInvalidGrant, "unknown quantity suffix in %q", s)
}
