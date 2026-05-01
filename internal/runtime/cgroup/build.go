package cgroup

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cruciblehq/crux/internal/crex"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/cruciblehq/crux/internal/manifest/grant"
	"github.com/cruciblehq/crux/internal/runtime/shared"
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
	oci      *specs.LinuxResources // Write target.
	internal *spec                 // Private conflict tracker.
	applied  map[string][]string   // Per-knob raw value strings, in insertion order, used to rebuild Unified.
}

// Returns a Subsystem that writes into resources as grants accumulate.
//
// resources is mutated in place by Build and Merge; callers retain ownership
// and observe updates without further coordination. The returned Subsystem
// also holds a private typed model used solely for v2-aware conflict
// detection; the model never escapes the package.
func New(resources *specs.LinuxResources) *Subsystem {
	return &Subsystem{
		oci:      resources,
		internal: newSpec(),
		applied:  make(map[string][]string),
	}
}

// Returns the cgroup subsystem identifier used by the runtime registry.
func (s *Subsystem) Name() shared.Name {
	return shared.NameCgroup
}

// Applies a parsed grant to the OCI write target.
//
// The grant has the form ".cgroup KNOB VALUES..." where KNOB is the dotted
// cgroup v2 knob name (e.g. cpu.weight, io.max) and VALUES are the
// knob-specific positional and keyword arguments. The grant is first
// validated against the private typed model, which enforces v2 semantics
// and rejects conflicting scalar grants. On success, the grant is projected
// into the OCI write target.
func (s *Subsystem) Build(g grant.Grant) error {
	if err := check(&g); err != nil {
		return err
	}
	knob, value, err := parseGrant(&g)
	if err != nil {
		return err
	}
	return s.projectKnob(knob, value)
}

// Folds the cgroup section of src into the wired-in section.
//
// Every entry of src.OCI.Linux.Resources.Unified is fed back through the
// per-grant projection path so the private conflict tracker stays accurate;
// src's device entries are also re-projected through the same path.
func (s *Subsystem) Merge(src shared.Spec) error {
	if src.OCI == nil || src.OCI.Linux == nil || src.OCI.Linux.Resources == nil {
		return nil
	}
	for knob, value := range src.OCI.Linux.Resources.Unified {
		if err := s.replayUnifiedKnob(knob, value); err != nil {
			return err
		}
	}
	for _, d := range src.OCI.Linux.Resources.Devices {
		if !d.Allow {
			continue // Deny entries are part of the deny-all baseline, not a grant.
		}
		if err := s.projectDeviceEntry(d); err != nil {
			return err
		}
	}
	return nil
}

// Replays the kernel-format value for one Unified knob through projectKnob.
//
// A single Unified entry may carry multiple kernel-format lines for list-
// typed knobs; each non-empty line is fed back through projectKnob so the
// typed model observes one entry per call.
func (s *Subsystem) replayUnifiedKnob(knob, value string) error {
	for line := range strings.SplitSeq(value, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if err := s.projectKnob(knob, line); err != nil {
			return err
		}
	}
	return nil
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
	if knob == devicesKnob || strings.HasPrefix(knob, devicesKnob+".") {
		s.flushDevices()
		return nil
	}
	existing := s.applied[knob]
	if !slices.Contains(existing, value) {
		s.applied[knob] = append(existing, value)
	}
	s.flushUnified(knob)
	return nil
}

// Re-projects a single OCI device entry through the per-grant path.
//
// Synthesises the textual "type major minor access" form expected by the
// typed model and dispatches through projectKnob so the conflict tracker and
// the OCI write target stay in step. An empty Type defaults to "a" (all),
// matching OCI semantics; missing Major or Minor default to 0.
func (s *Subsystem) projectDeviceEntry(d specs.LinuxDeviceCgroup) error {
	dt := d.Type
	if dt == "" {
		dt = "a"
	}
	var major, minor int64
	if d.Major != nil {
		major = *d.Major
	}
	if d.Minor != nil {
		minor = *d.Minor
	}
	value := fmt.Sprintf("%s %d %d %s", dt, major, minor, d.Access)
	return s.projectKnob(devicesKnob, value)
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
	s.oci.Devices = out
}

// Rewrites the Unified entry for knob from its per-knob applied list.
//
// Other knob entries in oci.Unified are untouched; this is called once per
// knob whenever its applied list changes. List-typed knobs become
// newline-separated kernel-format lines so each entry survives as a
// distinct value.
func (s *Subsystem) flushUnified(knob string) {
	if s.oci.Unified == nil {
		s.oci.Unified = make(map[string]string)
	}
	s.oci.Unified[knob] = strings.Join(s.applied[knob], "\n")
}

// Validates the surface shape of a cgroup grant.
//
// Rejects grants with a where clause (cgroup grants are unconditional), grants
// with no knob name, and grants with no value at all (no positional after the
// knob and no kwargs). Per-knob value validation runs later in the dispatcher.
func check(g *grant.Grant) error {
	if g.Where != nil {
		return crex.Wrapf(ErrInvalidGrant, "unexpected where clause in cgroup expression")
	}
	if len(g.Args) == 0 {
		return crex.Wrapf(ErrInvalidGrant, "missing knob name in cgroup expression")
	}
	if len(g.Args) < 2 && len(g.Kwargs) == 0 {
		return crex.Wrapf(ErrInvalidGrant, "missing value for knob in cgroup expression")
	}
	return nil
}

// Returns the grant's knob name and the reconstructed textual value.
//
// The knob name is g.Args[0], which must be of type ArgName; otherwise
// returns ErrInvalidGrant. The value is buildValue applied to the remaining
// positional and keyword arguments.
func parseGrant(g *grant.Grant) (string, string, error) {
	knobArg := g.Args[0]
	if knobArg.Type != grant.ArgName {
		return "", "", crex.Wrapf(ErrInvalidGrant, "expected name as knob in cgroup expression")
	}
	return knobArg.Value, buildValue(g.Args[1:], g.Kwargs), nil
}

// Joins positional arguments and key=value keyword arguments with single spaces.
//
// Produces the kernel-format value string that per-knob parsers consume.
// Positional arguments come first in declaration order, then keyword
// arguments in declaration order rendered as key=value.
func buildValue(args []grant.Arg, kwargs []grant.Kwarg) string {
	parts := make([]string, 0, len(args)+len(kwargs))
	for _, a := range args {
		parts = append(parts, a.Value)
	}
	for _, k := range kwargs {
		parts = append(parts, k.Key+"="+k.Value.Value)
	}
	return strings.Join(parts, " ")
}
