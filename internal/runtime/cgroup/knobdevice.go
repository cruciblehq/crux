package cgroup

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/internal/crex"
)

// Canonical access-mode order used when normalizing device access strings.
const deviceCanonicalAccessOrder = "rwm"

// Single device access permission entry in the cgroup v2 device controller.
type device struct {
	Type   deviceType `knob:"type" json:"type,omitempty"`     // Device type (character, block, or all).
	Major  uint32     `knob:"major" json:"major,omitempty"`   // Device major number.
	Minor  uint32     `knob:"minor" json:"minor,omitempty"`   // Device minor number.
	Access string     `knob:"access" json:"access,omitempty"` // Access permissions for the device.
}

// Device type selector for cgroup device controller entries.
type deviceType string

const (
	deviceTypeChar  deviceType = "c" // Character device.
	deviceTypeBlock deviceType = "b" // Block device.
	deviceTypeAll   deviceType = "a" // All device types (wildcard).
)

// Parses a device type character.
func parseDeviceType(value string) (deviceType, error) {
	s := strings.TrimSpace(value)
	switch deviceType(s) {
	case deviceTypeChar, deviceTypeBlock, deviceTypeAll:
		return deviceType(s), nil
	default:
		return "", crex.Wrapf(ErrInvalidGrant, "invalid device type %q", value)
	}
}

// Matches a device entry as "type major minor access". Type is any
// non-whitespace token (validated separately). Access must be a non-
// empty combination of r, w, m. For example, "c 8 0 rw" would specify
// a character device with major 8, minor 0, and read and write access.
var reDevice = regexp.MustCompile(`^(\S+)\s+(\d+)\s+(\d+)\s+([rwm]+)$`)

// Parses a device entry.
//
// Expects four whitespace-separated fields: type, major, minor, access. Type
// is c (character), b (block), or a (wildcard for both). Major and minor are
// unsigned integers. Access is a non-empty combination of r, w, and m in any
// order; the canonical deviceCanonicalAccessOrder order is applied when
// entries for the same device are merged by unioning their access bits.
func parseDevice(value string) (device, error) {
	m := reDevice.FindStringSubmatch(strings.TrimSpace(value))
	if m == nil {
		return device{}, crex.Wrapf(ErrInvalidGrant, "expected type major minor access")
	}
	maj, err := strconv.ParseUint(m[2], 10, 32)
	if err != nil {
		return device{}, crex.Wrap(ErrInvalidGrant, err)
	}
	min, err := strconv.ParseUint(m[3], 10, 32)
	if err != nil {
		return device{}, crex.Wrap(ErrInvalidGrant, err)
	}
	dt, err := parseDeviceType(m[1])
	if err != nil {
		return device{}, err
	}
	return device{
		Type:   dt,
		Major:  uint32(maj),
		Minor:  uint32(min),
		Access: m[4],
	}, nil
}

// Whether e and other address the same device (type, major, minor).
//
// Two device entries are considered to have the same identity if they specify
// the same device type, major number, and minor number, regardless of access
// permissions.
func (e device) equal(other device) bool {
	return e.Type == other.Type && e.Major == other.Major && e.Minor == other.Minor
}

// Always returns nil; device entries are always compatible.
//
// Devices with the same type, major, and minor number are merged by unioning
// their access permissions, so conflicting payloads cannot occur.
func (e device) check(other device) error {
	return nil
}

// Unions other's access bits into e and reports whether the access changed.
//
// For example, if e.Access is "rw" and other.Access is "rm", e.Access would
// be updated to "rwm" and the method would return true. If e.Access were
// already "rwm", the method would return false since no change was needed.
func (e *device) merge(other device) bool {
	if !e.equal(other) {
		return false
	}
	before := e.Access
	e.Access = mergeDeviceAccess(e.Access, other.Access)
	return e.Access != before
}

// Unions two device access strings, preserving canonical "rwm" order.
//
// For example, "rw" merged with "rm" would yield "rwm". The input strings
// must be non-empty and contain only characters in "rwm". The output string
// will contain the same characters as the union of the inputs, sorted in
// "rwm" order.
func mergeDeviceAccess(a, b string) string {
	var buf [3]byte
	n := 0
	for _, c := range deviceCanonicalAccessOrder {
		if strings.ContainsRune(a, c) || strings.ContainsRune(b, c) {
			buf[n] = byte(c)
			n++
		}
	}
	return string(buf[:n])
}
