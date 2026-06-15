package kernel

import (
	"fmt"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
	"github.com/cruciblehq/crux/crex"
)

// Implementation of the kernel requirement subsystem.
//
// Holds a pointer to the [Spec] passed at construction time. Each Build call
// writes one requirement to a specific field based on the grant.
type Subsystem struct {
	spec *Spec // pointer to the kernel spec to accumulate into.
}

// Returns a Subsystem wired to accumulate into spec.
func New(spec *Spec) *Subsystem {
	return &Subsystem{spec: spec}
}

// Returns the kernel subsystem identifier.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameKernel
}

// Returns the deduplication key for a kernel grant.
//
// The key is "type:value", ensuring two grants for the same type and value
// within one affordance are treated as a conflict.
func (s *Subsystem) Key(g *agl.Model) string {
	if len(g.Args) <= argValue {
		return ""
	}
	return fmt.Sprintf("%s:%s", g.Args[argType].Value, g.Args[argValue].Value)
}

// Applies a parsed grant to the kernel spec.
//
// The grant has the form ".kernel TYPE VALUE". TYPE selects the verification
// dimension; VALUE is the requirement to record. The accepted types are config
// (CONFIG_* flag without prefix), module (kernel module name), version (minimum
// kernel version), boot (boot parameter token), lsm (Linux Security Module name),
// hw (CPU hardware feature flag). No kwargs or where clause are accepted.
func (s *Subsystem) Build(g *agl.Model) error {
	if err := check(g); err != nil {
		return err
	}
	typ := g.Args[argType].Value
	value := g.Args[argValue].Value
	switch typ {
	case typeConfig:
		s.spec.Features = append(s.spec.Features, value)
	case typeModule:
		s.spec.Modules = append(s.spec.Modules, value)
	case typeVersion:
		s.spec.Versions = append(s.spec.Versions, value)
	case typeBoot:
		s.spec.BootParams = append(s.spec.BootParams, value)
	case typeLSM:
		s.spec.LSMs = append(s.spec.LSMs, value)
	case typeHW:
		s.spec.HWFeatures = append(s.spec.HWFeatures, value)
	default:
		return crex.Newf(ErrInvalidGrant, "unknown kernel grant type %q (expected config, module, version, boot, lsm, or hw)", typ)
	}
	return nil
}

// Validates the structural shape of a kernel grant.
func check(g *agl.Model) error {
	if g.Where != nil {
		return crex.Newf(ErrInvalidGrant, "unexpected where clause in kernel grant")
	}
	if len(g.Kwargs) > 0 {
		return crex.Newf(ErrInvalidGrant, "unexpected keyword arguments in kernel grant")
	}
	if len(g.Args) != kernelArgCount {
		return crex.Newf(ErrInvalidGrant, "kernel grant requires exactly two arguments (type and value)")
	}
	if g.Args[argType].Type != agl.ArgName {
		return crex.Newf(ErrInvalidGrant, "first argument must be a type name")
	}
	if g.Args[argValue].Type != agl.ArgName && g.Args[argValue].Type != agl.ArgStrASCII {
		return crex.Newf(ErrInvalidGrant, "second argument must be a name or quoted string value")
	}
	return nil
}
