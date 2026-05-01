package cgroup

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/internal/crex"
)

// Tag key used to mark cgroup spec fields as knobs.
//
// The build logic walks the spec struct to find the field tagged with a knob
// path and dispatches to the knob-specific parser. The merge logic uses the
// tag to identify list fields and merge entries by their identity keys.
const knobStructTag = "knob"

// Walks the spec struct to apply value at the field whose knob tag (or chain
// of knob-tagged ancestors) matches the dotted knob path.
//
// Lazily initialises the per-knob set tracker so repeated grants to the same
// scalar knob can be detected by checkScalarConflict. Returns ErrUnknownKnob
// when no field matches, ErrInvalidGrant on parse failure, ErrConflict on a
// disagreeing repeat assignment.
func (s *spec) applyKnob(knob, value string) error {
	if s.set == nil {
		s.set = make(map[string]struct{})
	}
	return applyStruct(s, reflect.TypeFor[spec](), reflect.ValueOf(s).Elem(), knob, knob, value)
}

// Dispatches the knob path to a matching field of t, recursing through
// untagged struct fields.
//
// remaining is the unconsumed portion of the dotted knob path; fullPath is
// the original path used only for error messages. Returns ErrUnknownKnob
// when no field of t (transitively) claims the path.
func applyStruct(s *spec, t reflect.Type, v reflect.Value, remaining string, fullPath string, value string) error {
	for i := range t.NumField() {
		handled, err := applyField(s, t.Field(i), v.Field(i), remaining, fullPath, value)
		if handled {
			return err
		}
	}
	return crex.Wrapf(ErrUnknownKnob, "unknown cgroup knob %q", fullPath)
}

// Tries to apply value via a single struct field.
//
// Returns handled=true when the field claims the knob path, in which case
// err is the result of applying value to that field (nil on success).
// Returns handled=false to let the caller continue searching sibling fields.
// Untagged struct fields are recursed into transparently and only count as
// handled when their subtree claims the path.
func applyField(s *spec, f reflect.StructField, fv reflect.Value, remaining, fullPath, value string) (bool, error) {
	tag := f.Tag.Get(knobStructTag)
	if tag != "" {
		return applyTaggedField(s, f, fv, tag, remaining, fullPath, value)
	}
	if f.Type.Kind() != reflect.Struct {
		return false, nil
	}
	err := applyStruct(s, f.Type, fv, remaining, fullPath, value)
	if errors.Is(err, ErrUnknownKnob) {
		return false, nil
	}
	return true, err
}

// Whether t is a slice that should be routed through applyListField.
//
// Includes all slices except indexList, which is parsed as a single scalar
// value by parseKindScalar.
func isListField(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && t != reflect.TypeFor[indexList]()
}

// Applies value when the field's knob tag matches or prefixes remaining.
//
// An exact tag match dispatches by field shape: scalar fields go through
// applyScalarField, list fields through applyListField. A prefix match
// strips the consumed segment plus the dot separator and recurses into the
// nested struct. List fields also claim a prefix-matching path so that
// composite leaf knobs (for example dmem.max) can route to applyListField
// with the original full path preserved for parser dispatch.
func applyTaggedField(s *spec, f reflect.StructField, fv reflect.Value, tag, remaining, fullPath, value string) (bool, error) {
	if tag == remaining {
		if isListField(f.Type) {
			return true, applyListField(s, fv, fullPath, value)
		}
		return true, applyScalarField(s, fv, fullPath, value)
	}
	if !strings.HasPrefix(remaining, tag+".") {
		return false, nil
	}
	suffix := remaining[len(tag)+1:]
	switch {
	case f.Type.Kind() == reflect.Struct:
		return true, applyStruct(s, f.Type, fv, suffix, fullPath, value)
	case isListField(f.Type):
		return true, applyListField(s, fv, fullPath, value)
	}
	return false, nil
}

// Parses value, rejects a conflicting repeat, then writes the field.
//
// On a normal scalar write the parsed value is assigned to fv and knob is
// recorded in s.set so a later grant to the same knob can be diffed by
// checkScalarConflict. io.weight is an exception: the same knob name accepts
// a scalar weight that belongs in fv and per-device entries that belong in
// s.IO.WeightDevices. normalizeIOWeight discriminates between the two: per-
// device entries are merged on the spot and the function returns without
// touching fv; scalar weights are returned in canonical numeric form for
// the rest of this function to parse and assign. See normalizeIOWeight for
// more details.
func applyScalarField(s *spec, fv reflect.Value, knob string, value string) error {
	if knob == ioWeightKnob {
		handled, normalised, err := normalizeIOWeight(s, value)
		if handled {
			return err
		}
		value = normalised
	}
	parsed, err := parseScalarValue(fv.Type(), value)
	if err != nil {
		return err
	}
	if err := checkScalarConflict(s, fv, knob, parsed); err != nil {
		return err
	}
	fv.Set(parsed)
	s.set[knob] = struct{}{}
	return nil
}

// Disambiguates the two value shapes accepted by the io.weight knob.
//
// io.weight accepts either a scalar weight (e.g., "100", "default 100") that
// sets the cgroup's overall IO weight, or a an entry (e.g., "8:16 200") that
// overrides the weight for one device. Both share the same knob name in the
// kernel, but map to different fields on spec: the scalar to a single field,
// per-device entries to s.IO.WeightDevices. For a per-device value, the entry
// is merged into s.IO.WeightDevices and (true, "", err) is returned so the
// caller stops. For a scalar value, the tuple (false, canonical, nil) is
// returned with the weight formatted as a plain decimal string; the caller
// continues by parsing canonical and writing it to the scalar field.
func normalizeIOWeight(s *spec, value string) (bool, string, error) {
	parts := strings.Fields(value)
	if len(parts) > 0 && parts[0] != "default" && strings.IndexByte(parts[0], ':') != -1 {
		wd, err := parseIOWeightDevice(value)
		if err != nil {
			return true, "", err
		}
		_, err = merge(&s.IO.WeightDevices, wd)
		return true, "", err
	}
	w, err := parseIOWeightScalar(parts)
	if err != nil {
		return false, "", err
	}
	return false, strconv.FormatUint(uint64(w), 10), nil
}

// Parses value as one list entry and merges it into the matching list slice on s.
//
// The element's type selects the per-knob parser and destination slice. The
// parsed entry is folded into the slice by merge, which appends, merges in
// place, or returns ErrConflict depending on the element's equal/check/merge
// methods. Returns ErrInvalidGrant for unknown element types.
func applyListField(s *spec, fv reflect.Value, knob string, value string) error {
	elemType := fv.Type().Elem()
	switch elemType {
	case reflect.TypeFor[hugeTLB]():
		return parseAndMerge(&s.HugeTLB, parseHugeTLB, value)
	case reflect.TypeFor[rdma]():
		return parseAndMerge(&s.RDMA, parseRDMA, value)
	case reflect.TypeFor[misc]():
		return parseAndMerge(&s.Misc, parseMisc, value)
	case reflect.TypeFor[device]():
		return parseAndMerge(&s.Devices, parseDevice, value)
	case reflect.TypeFor[ioMax]():
		return parseAndMerge(&s.IO.Max, parseIOMax, value)
	case reflect.TypeFor[ioLatency]():
		return parseAndMerge(&s.IO.Latency, parseIOLatency, value)
	case reflect.TypeFor[ioCost]():
		return parseAndMerge(&s.IO.Cost, parseIOCost, value)
	case reflect.TypeFor[ioCostQoS]():
		return parseAndMerge(&s.IO.CostQoS, parseIOCostQoS, value)
	case reflect.TypeFor[ioWeightDevice]():
		return parseAndMerge(&s.IO.WeightDevices, parseIOWeightDevice, value)
	case reflect.TypeFor[controller]():
		return parseAndMergeFunc(parseSubtreeControl, s.mergeSubtreeControl, value)
	case reflect.TypeFor[dmem]():
		return parseAndMerge(&s.Dmem, func(v string) (dmem, error) {
			return parseDmemEntry(knob, v)
		}, value)
	case reflect.TypeFor[psiTrigger]():
		return parseAndMergeFunc(parsePSITrigger, func(t psiTrigger) (bool, error) {
			return s.mergePSITriggers(knob, []psiTrigger{t})
		}, value)
	}
	return crex.Wrapf(ErrInvalidGrant, "unsupported list field type %s", elemType)
}

// Parses one list entry and folds it into dst.
//
// Used by applyListField for the common case where the destination is a plain
// *[]T on spec and the merge semantics come from T's mergeable methods.
func parseAndMerge[T any, PT mergeable[T]](dst *[]T, parse func(string) (T, error), value string) error {
	entry, err := parse(value)
	if err != nil {
		return err
	}
	_, err = merge[T, PT](dst, entry)
	return err
}

// Parses one list entry and feeds it to a custom merge callback.
//
// Used by applyListField for the cases that cannot use the standard merge,
// either because the merge needs additional context (the knob name for PSI
// triggers) or because the destination is not a flat *[]T (subtree_control
// merges into a controller bitset on spec, not a slice).
func parseAndMergeFunc[T any](parse func(string) (T, error), merge func(T) (bool, error), value string) error {
	entry, err := parse(value)
	if err != nil {
		return err
	}
	_, err = merge(entry)
	return err
}

// Parses value into a reflect.Value of t, dispatching by named type then by kind.
//
// Wraps parse failures with ErrInvalidGrant. ErrInvalidGrant is returned for
// unsupported types rather than panicking, since the dispatcher is reflective
// and may encounter arbitrary field types that it should reject gracefully.
func parseScalarValue(t reflect.Type, value string) (reflect.Value, error) {
	if v, ok, err := parseNamedScalar(t, value); ok {
		return v, err
	}
	return parseKindScalar(t, value)
}

// Parses values for named string-backed types whose set is constrained.
//
// Returns ok=true when t is one of the known named types so the caller does
// not fall through to the generic kind-based path. The reflect.Value is
// always of type t when ok=true, even on parse error.
func parseNamedScalar(t reflect.Type, value string) (reflect.Value, bool, error) {
	switch t {
	case reflect.TypeFor[nodeType]():
		v, err := parseNodeType(value)
		return reflect.ValueOf(v), true, err
	case reflect.TypeFor[partition]():
		v, err := parseCPUSetPartition(value)
		return reflect.ValueOf(v), true, err
	case reflect.TypeFor[ioPrioClass]():
		v, err := parseIOPrioClass(value)
		return reflect.ValueOf(v), true, err
	}
	return reflect.Value{}, false, nil
}

// Parses primitive-kinded types and the indexList slice into a reflect.Value of t.
//
// Numeric types use strconv with the type's bit width to reject overflow.
// indexList is recognised by exact type identity since other slice types
// belong to the list-field path. Wraps strconv errors with ErrInvalidGrant.
func parseKindScalar(t reflect.Type, value string) (reflect.Value, error) {
	zero := reflect.New(t).Elem()
	switch t.Kind() {
	case reflect.Bool:
		var v bool
		if err := parseBool(&v, value); err != nil {
			return reflect.Value{}, err
		}
		zero.SetBool(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(value, 10, t.Bits())
		if err != nil {
			return reflect.Value{}, crex.Wrap(ErrInvalidGrant, err)
		}
		zero.SetUint(n)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, t.Bits())
		if err != nil {
			return reflect.Value{}, crex.Wrap(ErrInvalidGrant, err)
		}
		zero.SetInt(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(value, t.Bits())
		if err != nil {
			return reflect.Value{}, crex.Wrap(ErrInvalidGrant, err)
		}
		zero.SetFloat(n)
	case reflect.Slice:
		if t == reflect.TypeFor[indexList]() {
			var v indexList
			if err := parseIndexList(&v, value); err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(v), nil
		}
		return reflect.Value{}, crex.Wrapf(ErrInvalidGrant, "unsupported slice type %s", t)
	default:
		return reflect.Value{}, crex.Wrapf(ErrInvalidGrant, "unsupported field type %s", t)
	}
	return zero, nil
}

// Whether knob has not been seen in this Build session.
//
// Returns nil when knob has not been seen in this Build session, or when the
// prior and incoming reflect.Values compare equal under reflect.Value.Equal
// (so idempotent repeats are accepted silently).
func checkScalarConflict(s *spec, fv reflect.Value, knob string, parsed reflect.Value) error {
	if _, seen := s.set[knob]; !seen {
		return nil
	}
	if fv.Equal(parsed) {
		return nil
	}
	return crex.Wrapf(ErrConflict, "%q already set to a different value", knob)
}
