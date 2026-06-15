package cgroup

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/crex"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
)

// Implementation for cgroup v2 resource limits.
//
// Holds the OCI Linux.Resources write target and a private typed model used
// solely for v2-aware conflict detection. Every successful grant is projected
// into the OCI write target: device knobs go to the typed Devices array (the
// v2 BPF program is synthesised from it by the runtime), every other knob goes
// to the Unified map verbatim in v2 kernel format. The private typed model
// never escapes the package.
type Subsystem struct {
	spec     *specs.LinuxResources // Write target.
	internal *spec                 // Private conflict tracker.
	applied  map[string][]string   // Per-knob raw value strings, in insertion order, used to rebuild Unified.
}

// Returns a Subsystem that writes into spec as grants accumulate.
//
// spec is mutated in place by Build and Merge; callers retain ownership and
// observe updates without further coordination. The returned Subsystem also
// holds a private typed model used solely for v2-aware conflict detection;
// the model never escapes the package.
func New(spec *specs.LinuxResources) *Subsystem {
	return &Subsystem{
		spec:     spec,
		internal: newSpec(),
		applied:  make(map[string][]string),
	}
}

// Returns the cgroup subsystem identifier used by the runtime registry.
func (s *Subsystem) Name() subsystem.Name {
	return subsystem.NameCgroup
}

// Returns the deduplication key for the grant.
//
// This key uniquely identifies the resource targeted by the grant. For scalar
// knobs (cpu.max, memory.max, and pids.max), the key is the knob name itself.
// For io.weight, the key is the knob name when the value is a plain weight, or
// io.weight:<major>:<minor> when the value targets a specific device. For IO
// list knobs addressed by a device pair (io.max, io.latency, io.cost.model,
// and io.cost.qos), the key is the knob name followed by the major and minor
// numbers separated by colons. For device knobs, the key consists of the knob
// name, device type, major number, and minor number separated by colons. For
// single-token identity knobs such as hugetlb, rdma.max, misc.max, dmem.*,
// cpu.pressure, memory.pressure, and io.pressure, the key is the knob name
// followed by the first value argument separated by a colon. Returns an empty
// string if no knob name is specified.
func (s *Subsystem) Key(g *agl.Model) string {
	if len(g.Args) == 0 {
		return ""
	}
	knob := g.Args[0].Value

	// Device knobs: identity is type + major + minor. Two grants for the same
	// device are a conflict; callers must express the complete access set in a
	// single grant.
	if knob == devicesKnob || strings.HasPrefix(knob, fmt.Sprintf("%s.", devicesKnob)) {
		if len(g.Args) >= 4 {
			return fmt.Sprintf("%s:%s:%s:%s", knob, g.Args[1].Value, g.Args[2].Value, g.Args[3].Value)
		}
		return knob // too few args; Build will produce the error
	}

	// List knobs addressed by a MAJOR MINOR pair (two separate positional args).
	switch knob {
	case ioMaxKnob, ioLatencyKnob, ioCostModelKnob, ioCostQoSKnob:
		if len(g.Args) >= 3 {
			return fmt.Sprintf("%s:%s:%s", knob, g.Args[1].Value, g.Args[2].Value)
		}
		return knob // too few args; Build will produce the error
	}

	// io.weight is scalar when its value arg is a plain number, and per-device
	// when the value arg is a "major:minor" token (which contains ":").
	if knob == ioWeightKnob {
		if len(g.Args) >= 2 && strings.Contains(g.Args[1].Value, ":") {
			return fmt.Sprintf("%s:%s", knob, g.Args[1].Value)
		}
		return knob
	}

	// List knobs addressed by a single name or type token (first value arg):
	// hugetlb (page size), rdma.max/misc.max (device/resource name), dmem.*
	// (region name), cpu.pressure/memory.pressure/io.pressure (PSI event type).
	switch knob {
	case hugeTLBKnob, rdmaKnob, miscKnob,
		dmemMaxKnob, dmemMinKnob, dmemLowKnob,
		psiCPUKnob, psiMemoryKnob, psiIOKnob:
		if len(g.Args) >= 2 {
			return fmt.Sprintf("%s:%s", knob, g.Args[1].Value)
		}
		return knob // too few args; Build will produce the error
	}

	// All remaining knobs are scalar: the knob name alone is the key.
	return knob
}

// Applies a parsed grant to the OCI write target.
//
// The grant has the form ".cgroup KNOB VALUES..." where KNOB is the dotted
// cgroup v2 knob name (e.g. cpu.weight) and VALUES are the knob-specific
// positional and keyword arguments. The grant is first validated against the
// private typed model, which enforces v2 semantics and rejects conflicting
// scalar grants. On success, the grant is projected into the OCI write target.
func (s *Subsystem) Build(g *agl.Model) error {
	if err := check(g); err != nil {
		return err
	}
	knob, value, err := parseGrant(g)
	if err != nil {
		return err
	}
	return s.projectKnob(knob, value)
}

// Projects a knob/value pair through the typed model into the OCI write target.
//
// Validates and merges the value into the typed model first, so structural,
// catalog, and same-knob conflict checks all run before any OCI mutation.
// Device knobs flush the typed Devices array; every other knob appends value
// to the per-knob applied list (deduplicated by exact string equality so
// idempotent scalar replays do not accumulate Unified lines) and then rebuilds
// Unified[knob] from that list.
func (s *Subsystem) projectKnob(knob, value string) error {
	if err := s.internal.applyKnob(knob, value); err != nil {
		return err
	}
	if knob == devicesKnob || strings.HasPrefix(knob, fmt.Sprintf("%s.", devicesKnob)) {
		s.flushDevices()
		return nil
	}
	if knob == hugeTLBKnob {
		s.flushHugeTLB()
		return nil
	}
	existing := s.applied[knob]
	if !slices.Contains(existing, value) {
		s.applied[knob] = append(existing, value)
	}
	s.flushUnified(knob)
	return nil
}

// Rewrites oci.Devices from the typed model as a positive allowlist.
//
// Every emitted entry has Allow=true; the deny-everything baseline lives
// outside the cgroup spec, so the typed model only carries grants and the
// OCI device list is rebuilt to mirror them.
func (s *Subsystem) flushDevices() {
	out := make([]specs.LinuxDeviceCgroup, 0, len(s.internal.Devices))
	for _, d := range s.internal.Devices {
		major := int64(d.Major)
		minor := int64(d.Minor)
		out = append(out, specs.LinuxDeviceCgroup{
			Allow:  true,
			Type:   string(d.Type),
			Major:  &major,
			Minor:  &minor,
			Access: d.Access,
		})
	}
	s.spec.Devices = out
}

// Rewrites the Unified entries for all accumulated huge page size limits.
//
// Each hugeTLB entry produces one or two Unified keys: hugetlb.<size>.max is
// always written; hugetlb.<size>.rsvd.max is written only when RsvdMax is
// non-zero. This is called whenever a hugeTLB grant is applied, so Unified
// always reflects the current state of the typed model.
func (s *Subsystem) flushHugeTLB() {
	if s.spec.Unified == nil {
		s.spec.Unified = make(map[string]string)
	}
	for _, h := range s.internal.HugeTLB {
		s.spec.Unified[fmt.Sprintf("hugetlb.%s.max", h.Size)] = strconv.FormatUint(h.Max, 10)
		if h.RsvdMax != 0 {
			s.spec.Unified[fmt.Sprintf("hugetlb.%s.rsvd.max", h.Size)] = strconv.FormatUint(h.RsvdMax, 10)
		}
	}
}

// Rewrites the Unified entry for knob from its per-knob applied list.
//
// Other knob entries in Unified are untouched; this is called once per knob
// whenever its applied list changes. List-typed knobs become newline-separated
// kernel-format lines so each entry survives as a distinct value.
func (s *Subsystem) flushUnified(knob string) {
	if s.spec.Unified == nil {
		s.spec.Unified = make(map[string]string)
	}
	s.spec.Unified[knob] = strings.Join(s.applied[knob], "\n")
}

// Validates the surface shape of a cgroup agl.
//
// Rejects grants with a where clause (cgroup grants are unconditional), grants
// with no knob name, and grants with no value at all (no positional after the
// knob and no kwargs). Per-knob value validation runs later in the dispatcher.
func check(g *agl.Model) error {
	if g.Where != nil {
		return crex.Newf(ErrInvalidGrant, "unexpected where clause in cgroup expression")
	}
	if len(g.Args) == 0 {
		return crex.Newf(ErrInvalidGrant, "missing knob name in cgroup expression")
	}
	if len(g.Args) < 2 && len(g.Kwargs) == 0 {
		return crex.Newf(ErrInvalidGrant, "missing value for knob in cgroup expression")
	}
	return nil
}

// Returns the grant's knob name and the reconstructed textual value.
//
// The knob name is g.Args[0], which must be of type ArgName; otherwise returns
// ErrInvalidGrant. The value is buildValue applied to the remaining positional
// and keyword arguments.
func parseGrant(g *agl.Model) (string, string, error) {
	knobArg := g.Args[0]
	if knobArg.Type != agl.ArgName {
		return "", "", crex.Newf(ErrInvalidGrant, "expected name as knob in cgroup expression")
	}
	return knobArg.Value, buildValue(g.Args[1:], g.Kwargs), nil
}

// Joins positional arguments and key=value keyword arguments with single spaces.
//
// Produces the kernel-format value string that per-knob parsers consume.
// Positional arguments come first in declaration order, then keyword
// arguments in declaration order rendered as key=value.
func buildValue(args []agl.Arg, kwargs []agl.Kwarg) string {
	parts := make([]string, 0, len(args)+len(kwargs))
	for _, a := range args {
		parts = append(parts, a.Value)
	}
	for _, k := range kwargs {
		parts = append(parts, fmt.Sprintf("%s=%s", k.Key, k.Value.Value))
	}
	return strings.Join(parts, " ")
}
