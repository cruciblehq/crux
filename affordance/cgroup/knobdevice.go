package cgroup

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/crex"
)

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

// Matches a device entry as "type major minor access".
//
// Type is any non-whitespace token (validated separately). Access must be a
// non-empty combination of r, w, m. For example, "c 8 0 rw" would specify a
// character device with major 8, minor 0, and read and write access.
var reDevice = regexp.MustCompile(`^(\S+)\s+(\d+)\s+(\d+)\s+([rwm]+)$`)

// Parses a device entry.
//
// Expects four whitespace-separated fields: type, major, minor, access. Type
// is c (character), b (block), or a (wildcard for both). Major and minor are
// unsigned integers. Access is a non-empty combination of r, w, and m in any
// order. The access string is stored exactly as provided; the caller must
// supply the complete intended access (e.g. "rw" for both read and write).
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

// Returns an error when e and other address the same device.
//
// Two grants for the same type, major, and minor conflict regardless of
// their access permissions. Callers must express all required access bits
// in a single grant (e.g. "rw" rather than two separate "r" and "w" grants).
func (e device) check(other device) error {
	if !e.equal(other) {
		return nil
	}
	return crex.Wrapf(ErrConflict, "%s %s %d:%d already granted", devicesKnob, other.Type, other.Major, other.Minor)
}

// Always returns false; same-identity device entries are rejected by check
// before merge is ever reached.
func (e *device) merge(_ device) bool {
	return false
}
