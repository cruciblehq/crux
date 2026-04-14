package subsystem

import (
	"strings"

	"github.com/cruciblehq/crex"
)

// Parses key-value args into struct fields via [setScalar].
//
// Each arg is "key value". The fields map binds argument keys to field
// pointers whose concrete type determines the parse function.
func parseArgs(args []string, fields map[knobArg]any) error {
	for _, arg := range args {
		key, val, _ := strings.Cut(arg, " ")
		val = strings.TrimSpace(val)
		ptr, ok := fields[knobArg(key)]
		if !ok {
			return crex.Wrapf(ErrGrantExpression, "unknown key %q", key)
		}
		if err := setScalar(ptr, val); err != nil {
			return err
		}
	}
	return nil
}

// Parses an io.max entry from args.
//
// Entries are keyed by major:minor. Identical entries are dropped.
// Conflicting entries for the same device error.
func (cg *cgroup) parseIOMax(args []string) (bool, error) {
	var m cgroupIOMax
	if err := parseArgs(args, map[knobArg]any{
		knobArgMajor: &m.Major,
		knobArgMinor: &m.Minor,
		knobArgRbps:  &m.Rbps,
		knobArgWbps:  &m.Wbps,
		knobArgRiops: &m.Riops,
		knobArgWiops: &m.Wiops,
	}); err != nil {
		return false, err
	}
	for _, e := range cg.IO.Max {
		if e.Major == m.Major && e.Minor == m.Minor {
			if e == m {
				return false, nil
			}
			return false, crex.Wrapf(ErrGrantConflict, "io.max %d:%d already set", m.Major, m.Minor)
		}
	}
	cg.IO.Max = append(cg.IO.Max, m)
	return true, nil
}

// Parses an io.latency entry from args.
//
// Entries are keyed by major:minor. Identical entries are dropped.
// Conflicting entries for the same device error.
func (cg *cgroup) parseIOLatency(args []string) (bool, error) {
	var l cgroupIOLatency
	if err := parseArgs(args, map[knobArg]any{
		knobArgMajor:  &l.Major,
		knobArgMinor:  &l.Minor,
		knobArgTarget: &l.Target,
	}); err != nil {
		return false, err
	}
	for _, e := range cg.IO.Latency {
		if e.Major == l.Major && e.Minor == l.Minor {
			if e == l {
				return false, nil
			}
			return false, crex.Wrapf(ErrGrantConflict, "io.latency %d:%d already set", l.Major, l.Minor)
		}
	}
	cg.IO.Latency = append(cg.IO.Latency, l)
	return true, nil
}

// Parses an io.cost.model entry from args.
//
// Entries are keyed by major:minor. Identical entries are dropped.
// Conflicting entries for the same device error.
func (cg *cgroup) parseIOCost(args []string) (bool, error) {
	var c cgroupIOCost
	if err := parseArgs(args, map[knobArg]any{
		knobArgMajor:     &c.Major,
		knobArgMinor:     &c.Minor,
		knobArgRbps:      &c.Rbps,
		knobArgWbps:      &c.Wbps,
		knobArgRseqiops:  &c.Rseqiops,
		knobArgRrandiops: &c.Rrandiops,
		knobArgWseqiops:  &c.Wseqiops,
		knobArgWrandiops: &c.Wrandiops,
	}); err != nil {
		return false, err
	}
	for _, e := range cg.IO.Cost {
		if e.Major == c.Major && e.Minor == c.Minor {
			if e == c {
				return false, nil
			}
			return false, crex.Wrapf(ErrGrantConflict, "io.cost.model %d:%d already set", c.Major, c.Minor)
		}
	}
	cg.IO.Cost = append(cg.IO.Cost, c)
	return true, nil
}

// Parses an io.cost.qos entry from args.
//
// Entries are keyed by major:minor. Identical entries are dropped.
// Conflicting entries for the same device error.
func (cg *cgroup) parseIOCostQoS(args []string) (bool, error) {
	var q cgroupIOCostQoS
	if err := parseArgs(args, map[knobArg]any{
		knobArgMajor: &q.Major,
		knobArgMinor: &q.Minor,
		knobArgRpct:  &q.Rpct,
		knobArgRlat:  &q.Rlat,
		knobArgWpct:  &q.Wpct,
		knobArgWlat:  &q.Wlat,
		knobArgMin:   &q.Min,
		knobArgMax:   &q.Max,
	}); err != nil {
		return false, err
	}
	for _, e := range cg.IO.CostQoS {
		if e.Major == q.Major && e.Minor == q.Minor {
			if e == q {
				return false, nil
			}
			return false, crex.Wrapf(ErrGrantConflict, "io.cost.qos %d:%d already set", q.Major, q.Minor)
		}
	}
	cg.IO.CostQoS = append(cg.IO.CostQoS, q)
	return true, nil
}

// Parses a hugetlb entry (positional size, args for max and rsvd_max).
//
// Entries are keyed by page size. Identical entries are dropped.
// Conflicting entries for the same size error.
func (cg *cgroup) parseHugeTLB(size string, args []string) (bool, error) {
	size = strings.TrimSpace(size)
	if size == "" {
		return false, crex.Wrapf(ErrGrantExpression, "page size required")
	}
	h := cgroupHugeTLB{Size: size}
	if err := parseArgs(args, map[knobArg]any{
		knobArgMax:     &h.Max,
		knobArgRsvdMax: &h.RsvdMax,
	}); err != nil {
		return false, err
	}
	for _, e := range cg.HugeTLB {
		if e.Size == h.Size {
			if e == h {
				return false, nil
			}
			return false, crex.Wrapf(ErrGrantConflict, "hugetlb %s already set", h.Size)
		}
	}
	cg.HugeTLB = append(cg.HugeTLB, h)
	return true, nil
}

// Parses an rdma entry (positional device, args for hca_handle and hca_object).
//
// Entries are keyed by device name. Identical entries are dropped.
// Conflicting entries for the same device error.
func (cg *cgroup) parseRDMA(device string, args []string) (bool, error) {
	device = strings.TrimSpace(device)
	if device == "" {
		return false, crex.Wrapf(ErrGrantExpression, "device name required")
	}
	r := cgroupRDMA{Device: device}
	if err := parseArgs(args, map[knobArg]any{
		knobArgHcaHandle: &r.HcaHandle,
		knobArgHcaObject: &r.HcaObject,
	}); err != nil {
		return false, err
	}
	for _, e := range cg.RDMA {
		if e.Device == r.Device {
			if e == r {
				return false, nil
			}
			return false, crex.Wrapf(ErrGrantConflict, "rdma %s already set", r.Device)
		}
	}
	cg.RDMA = append(cg.RDMA, r)
	return true, nil
}

// Parses a misc entry (positional resource, args for max).
//
// Entries are keyed by resource name. Identical entries are dropped.
// Conflicting entries for the same resource error.
func (cg *cgroup) parseMisc(resource string, args []string) (bool, error) {
	resource = strings.TrimSpace(resource)
	if resource == "" {
		return false, crex.Wrapf(ErrGrantExpression, "resource name required")
	}
	m := cgroupMisc{Resource: resource}
	if err := parseArgs(args, map[knobArg]any{
		knobArgMax: &m.Max,
	}); err != nil {
		return false, err
	}
	for _, e := range cg.Misc {
		if e.Resource == m.Resource {
			if e == m {
				return false, nil
			}
			return false, crex.Wrapf(ErrGrantConflict, "misc %s already set", m.Resource)
		}
	}
	cg.Misc = append(cg.Misc, m)
	return true, nil
}

// Parses a device entry and merges into the model.
//
// Entries are keyed by type, major, and minor. Access flags are unioned:
// a new grant adds permissions but never removes them. Returns false if
// the incoming access is already a subset of the existing entry.
func (cg *cgroup) parseDevice(val string) (bool, error) {
	d, err := parseDeviceEntry(val)
	if err != nil {
		return false, err
	}
	for i := range cg.Devices {
		e := &cg.Devices[i]
		if e.Type == d.Type && e.Major == d.Major && e.Minor == d.Minor {
			merged := mergeDeviceAccess(e.Access, d.Access)
			if merged == e.Access {
				return false, nil
			}
			e.Access = merged
			return true, nil
		}
	}
	cg.Devices = append(cg.Devices, d)
	return true, nil
}

// Unions two device access strings, preserving canonical "rwm" order.
func mergeDeviceAccess(a, b string) string {
	var buf [3]byte
	n := 0
	for _, c := range []deviceAccess{deviceRead, deviceWrite, deviceMknod} {
		if strings.ContainsRune(a, rune(c)) || strings.ContainsRune(b, rune(c)) {
			buf[n] = byte(c)
			n++
		}
	}
	return string(buf[:n])
}

// Parses a single device expression ("<type> <major> <minor> <access>").
func parseDeviceEntry(val string) (cgroupDevice, error) {
	fields := strings.Fields(val)
	if len(fields) != 4 {
		return cgroupDevice{}, crex.Wrapf(ErrGrantExpression, "invalid expression %q", val)
	}
	var d cgroupDevice
	var err error
	d.Type, err = parseDeviceKind(fields[0])
	if err != nil {
		return cgroupDevice{}, err
	}
	if err := parseUint32(&d.Major, fields[1]); err != nil {
		return cgroupDevice{}, err
	}
	if err := parseUint32(&d.Minor, fields[2]); err != nil {
		return cgroupDevice{}, err
	}
	d.Access = fields[3]
	for _, c := range d.Access {
		switch deviceAccess(c) {
		case deviceRead, deviceWrite, deviceMknod:
		default:
			return cgroupDevice{}, crex.Wrapf(ErrGrantExpression, "invalid access char %q in %q", string(c), d.Access)
		}
	}
	return d, nil
}

// Parses PSI triggers and merges into the destination slice.
//
// Entries are keyed by kind. Identical entries are dropped.
// Conflicting entries for the same kind error.
func parsePSI(dst *[]cgroupPSITrigger, args []string) (bool, error) {
	any := false
	for _, arg := range args {
		tr, err := parsePSITrigger(arg)
		if err != nil {
			return false, err
		}
		found := false
		for _, e := range *dst {
			if e.Kind == tr.Kind {
				if e == tr {
					found = true
					break
				}
				return false, crex.Wrapf(ErrGrantConflict, "psi %s already set", tr.Kind)
			}
		}
		if !found {
			*dst = append(*dst, tr)
			any = true
		}
	}
	return any, nil
}

// Parses a single PSI trigger expression ("<kind> <threshold> <window>").
func parsePSITrigger(arg string) (cgroupPSITrigger, error) {
	fields := strings.Fields(arg)
	if len(fields) != 3 {
		return cgroupPSITrigger{}, crex.Wrapf(ErrGrantExpression, "invalid expression %q", arg)
	}
	var tr cgroupPSITrigger
	var err error
	tr.Kind, err = parsePsiKind(fields[0])
	if err != nil {
		return cgroupPSITrigger{}, err
	}
	if err := parseUint64(&tr.Threshold, fields[1]); err != nil {
		return cgroupPSITrigger{}, err
	}
	if err := parseUint64(&tr.Window, fields[2]); err != nil {
		return cgroupPSITrigger{}, err
	}
	return tr, nil
}
