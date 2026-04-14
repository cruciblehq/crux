package subsystem

import (
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/cruciblehq/crex"
)

// Knob name as it appears in a grant expression.
type cgroupKnobName string

// Struct tag key used to mark cgroup model fields as scalar knobs.
const knobStructTag = "knob"

// Descriptor for a single cgroup knob.
type cgroupKnob struct {

	// True for list knobs that accept multiple grants.
	//
	// Scalar knobs that compare against a restrictive default return false for
	// no-ops and error on conflicts.
	list bool

	// Returns true if the grant had an effect on the model.
	//
	// Scalar knobs return false for no-ops (value equals the restrictive default
	// or matches the current value) and error on conflicts. List knobs always
	// return true on success.
	apply func(cg *cgroup, val string, args []string) (bool, error)
}

// Knob names for composite and custom-parsed interface files.
const (

	// I/O controller (per-device bandwidth, latency, and cost tuning).
	knobIOMax       cgroupKnobName = "io.max"        // Per-device bandwidth and IOPS caps.
	knobIOLatency   cgroupKnobName = "io.latency"    // Per-device latency targets.
	knobIOCostModel cgroupKnobName = "io.cost.model" // Per-device cost model coefficients.
	knobIOCostQoS   cgroupKnobName = "io.cost.qos"   // Per-device cost QoS parameters.

	// Composite resource controllers (keyed by device, page size, or name).
	knobHugeTLB cgroupKnobName = "hugetlb" // Huge page limits per page size.
	knobRDMA    cgroupKnobName = "rdma"    // RDMA resource limits per HCA device.
	knobMisc    cgroupKnobName = "misc"    // Miscellaneous scalar resources (e.g. SEV slots).
	knobDevice  cgroupKnobName = "device"  // Device access permissions (BPF-enforced).

	// Pressure stall information (contention triggers per resource).
	knobPSICPU    cgroupKnobName = "psi.cpu"    // CPU pressure triggers.
	knobPSIMemory cgroupKnobName = "psi.memory" // Memory pressure triggers.
	knobPSIIO     cgroupKnobName = "psi.io"     // I/O pressure triggers.

	// Core hierarchy control.
	knobSubtreeControl cgroupKnobName = "cgroup.subtree_control" // Controllers delegated to children.
)

// Argument key within a composite knob's sub-arguments.
type knobArg string

// Argument keys shared across multiple knob parsers.
const (
	knobArgMajor knobArg = "major" // Block device major number.
	knobArgMinor knobArg = "minor" // Block device minor number.
	knobArgRbps  knobArg = "rbps"  // Read bytes/sec.
	knobArgWbps  knobArg = "wbps"  // Write bytes/sec.
	knobArgMax   knobArg = "max"   // Maximum allocation.
)

// Argument keys for I/O bandwidth and IOPS caps (io.max).
const (
	knobArgRiops knobArg = "riops" // Max read IOPS.
	knobArgWiops knobArg = "wiops" // Max write IOPS.
)

// Argument keys for I/O latency targets (io.latency).
const (
	knobArgTarget knobArg = "target" // Latency target in microseconds.
)

// Argument keys for I/O cost model coefficients (io.cost.model).
const (
	knobArgRseqiops  knobArg = "rseqiops"  // Sequential read IOPS capacity.
	knobArgRrandiops knobArg = "rrandiops" // Random read IOPS capacity.
	knobArgWseqiops  knobArg = "wseqiops"  // Sequential write IOPS capacity.
	knobArgWrandiops knobArg = "wrandiops" // Random write IOPS capacity.
)

// Argument keys for I/O cost QoS parameters (io.cost.qos).
const (
	knobArgRpct knobArg = "rpct" // Read latency percentile.
	knobArgRlat knobArg = "rlat" // Read latency target in microseconds.
	knobArgWpct knobArg = "wpct" // Write latency percentile.
	knobArgWlat knobArg = "wlat" // Write latency target in microseconds.
	knobArgMin  knobArg = "min"  // Minimum weight fraction.
)

// Argument keys for huge page limits (hugetlb).
const (
	knobArgRsvdMax knobArg = "rsvd_max" // Reserved huge page maximum in bytes.
)

// Argument keys for RDMA resource limits (rdma).
const (
	knobArgHcaHandle knobArg = "hca_handle" // Max HCA handles.
	knobArgHcaObject knobArg = "hca_object" // Max HCA objects.
)

// Maps typed-string field types to their allowed values.
//
// Fields whose reflect.Type appears here are validated before assignment.
// The keys must match the "knob" struct tags on the corresponding fields.
var knobValidators = map[reflect.Type][]string{
	reflect.TypeFor[ioPrioClass]():     {string(ioPrioRT), string(ioPrioBE), string(ioPrioIdle)},
	reflect.TypeFor[cgroupPartition](): {string(cgroupPartitionMember), string(cgroupPartitionRoot), string(cgroupPartitionIsolated)},
	reflect.TypeFor[cgroupNodeType]():  {string(cgroupNodeDomain), string(cgroupNodeThreaded)},
}

// Extracts the "default=X" value from a field's codec struct tag.
//
// Returns the empty string when no default is declared.
func extractCodecDefault(f reflect.StructField) string {
	tag := f.Tag.Get("codec")
	for _, part := range strings.Split(tag, ",") {
		if v, ok := strings.CutPrefix(part, "default="); ok {
			return v
		}
	}
	return ""
}

// Parses val and assigns it to a struct field pointer.
//
// Dispatches on the concrete type of ptr. Typed string fields are validated
// against [knobValidators] when present.
func setScalar(ptr any, val string) error {
	switch p := ptr.(type) {
	case *uint64:
		return parseUint64(p, val)
	case *uint32:
		return parseUint32(p, val)
	case *uint16:
		return parseUint16(p, val)
	case *float64:
		return parseFloat64(p, val)
	case *bool:
		return parseBool(p, val)
	default:
		rv := reflect.ValueOf(ptr).Elem()
		crex.Assertf(rv.Kind() == reflect.String, "unsupported scalar type %T", ptr)
		if allowed, ok := knobValidators[rv.Type()]; ok {
			if !slices.Contains(allowed, val) {
				return crex.Wrapf(ErrGrantExpression, "invalid value %q", val)
			}
		}
		rv.SetString(val)
		return nil
	}
}

// Maps every recognized cgroup knob name to its descriptor.
//
// Scalar knobs are auto-registered from struct tags at init time.
// List and composite knobs are registered manually below.
var cgroupKnobs map[cgroupKnobName]cgroupKnob

// Knobs that need custom parsing.
//
// Covers list/composite knobs that append entries and scalars with non-trivial
// parse logic. The keys must match the "knob" struct tags on the corresponding
// struct fields.
var cgroupCustomKnobs = map[cgroupKnobName]cgroupKnob{

	// I/O controller.
	knobIOMax:       {list: true, apply: func(cg *cgroup, _ string, a []string) (bool, error) { return cg.parseIOMax(a) }},
	knobIOLatency:   {list: true, apply: func(cg *cgroup, _ string, a []string) (bool, error) { return cg.parseIOLatency(a) }},
	knobIOCostModel: {list: true, apply: func(cg *cgroup, _ string, a []string) (bool, error) { return cg.parseIOCost(a) }},
	knobIOCostQoS:   {list: true, apply: func(cg *cgroup, _ string, a []string) (bool, error) { return cg.parseIOCostQoS(a) }},

	// Composite resource controllers.
	knobHugeTLB: {list: true, apply: func(cg *cgroup, v string, a []string) (bool, error) { return cg.parseHugeTLB(v, a) }},
	knobRDMA:    {list: true, apply: func(cg *cgroup, v string, a []string) (bool, error) { return cg.parseRDMA(v, a) }},
	knobMisc:    {list: true, apply: func(cg *cgroup, v string, a []string) (bool, error) { return cg.parseMisc(v, a) }},
	knobDevice:  {list: true, apply: func(cg *cgroup, v string, _ []string) (bool, error) { return cg.parseDevice(v) }},

	// PSI triggers.
	knobPSICPU:    {list: true, apply: func(cg *cgroup, _ string, a []string) (bool, error) { return parsePSI(&cg.PSI.CPU, a) }},
	knobPSIMemory: {list: true, apply: func(cg *cgroup, _ string, a []string) (bool, error) { return parsePSI(&cg.PSI.Memory, a) }},
	knobPSIIO:     {list: true, apply: func(cg *cgroup, _ string, a []string) (bool, error) { return parsePSI(&cg.PSI.IO, a) }},

	// Core hierarchy control.
	knobSubtreeControl: {apply: func(cg *cgroup, v string, _ []string) (bool, error) {
		incoming := strings.Fields(v)
		slices.Sort(incoming)
		if len(incoming) == 0 {
			return false, nil
		}
		if slices.Equal(cg.Core.SubtreeControl, incoming) {
			return false, nil
		}
		if len(cg.Core.SubtreeControl) > 0 {
			return false, crex.Wrapf(ErrGrantConflict, "cgroup %s already set", knobSubtreeControl)
		}
		cg.Core.SubtreeControl = incoming
		return true, nil
	}},
}

// Populates [cgroupKnobs] from struct tags and custom definitions.
//
// Scalar knobs are derived from [cgroup] fields tagged with [knobStructTag].
// List and composite knobs are defined in [cgroupCustomKnobs].
func init() {
	cgroupKnobs = make(map[cgroupKnobName]cgroupKnob)
	registerScalarKnobs(reflect.TypeFor[cgroup](), nil, cgroupKnobs)
	maps.Copy(cgroupKnobs, cgroupCustomKnobs)
}

// Registers scalar knob parsers by walking the struct type recursively.
//
// Each field carrying a [knobStructTag] struct tag gets a parser that
// compares the incoming value against the field's restrictive default
// (from the codec tag) and the current model value to determine whether
// the grant is a no-op, a conflict, or takes effect.
func registerScalarKnobs(t reflect.Type, path []int, dst map[cgroupKnobName]cgroupKnob) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fp := append(slices.Clone(path), i)

		tag := f.Tag.Get(knobStructTag)
		if tag != "" {
			name := cgroupKnobName(tag)
			defaultVal := scalarDefault(f)
			dst[name] = cgroupKnob{apply: func(cg *cgroup, val string, _ []string) (bool, error) {
				field := reflect.ValueOf(cg).Elem().FieldByIndex(fp)
				tmp := reflect.New(field.Type())
				if err := setScalar(tmp.Interface(), val); err != nil {
					return false, err
				}
				return applyScalar(field, tmp.Elem(), defaultVal, name)
			}}
			continue
		}

		if f.Type.Kind() == reflect.Struct {
			registerScalarKnobs(f.Type, fp, dst)
		}
	}
}

// Computes the restrictive default for a scalar knob field.
//
// Returns the zero value when no "default=X" is declared in the codec tag.
// Panics if the declared default cannot be parsed into the field's type.
func scalarDefault(f reflect.StructField) reflect.Value {
	def := extractCodecDefault(f)
	if def == "" {
		return reflect.Zero(f.Type)
	}
	tmp := reflect.New(f.Type)
	if err := setScalar(tmp.Interface(), def); err != nil {
		panic("invalid default for " + f.Tag.Get(knobStructTag) + ": " + err.Error())
	}
	return tmp.Elem()
}

// Applies a scalar knob value using three-way comparison.
//
// Returns (false, nil) if the value equals the restrictive default (not a
// relaxation) or the current field value (idempotent). Returns an error
// wrapping [ErrGrantConflict] if the field is already relaxed to a different
// value. Otherwise sets the field and returns (true, nil).
func applyScalar(field, incoming, defaultVal reflect.Value, name cgroupKnobName) (bool, error) {
	if incoming.Equal(defaultVal) {
		return false, nil
	}
	if incoming.Equal(field) {
		return false, nil
	}
	if !field.Equal(defaultVal) {
		return false, crex.Wrapf(ErrGrantConflict, "cgroup %s already set", name)
	}
	field.Set(incoming)
	return true, nil
}
