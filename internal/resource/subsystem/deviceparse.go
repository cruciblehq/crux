package subsystem

import (
	"strings"

	"github.com/cruciblehq/crex"
)

// Parses a device provisioning grant from an expression and args.
//
// The expression is the device name (e.g., "null", "fuse", "nvidia0"). Args
// carry the device properties as "key value" pairs (path, type, major, minor,
// file_mode, uid, gid). Path, type, major, minor, and file_mode are required;
// uid and gid are optional (defaulting to nobody/nogroup).
func parseDeviceExpr(expr string, args []string) (device, error) {
	name := strings.TrimSpace(expr)
	if name == "" {
		return device{}, crex.Wrapf(ErrGrantExpression, "device name required")
	}
	if err := validateDeviceName(name); err != nil {
		return device{}, err
	}

	d := device{Name: name}
	seen := make(map[deviceArg]bool)
	for _, arg := range args {
		key, val, ok := strings.Cut(arg, " ")
		if !ok {
			return device{}, crex.Wrapf(ErrGrantExpression, "invalid arg %q", arg)
		}
		val = strings.TrimSpace(val)
		k := deviceArg(key)
		if seen[k] {
			return device{}, crex.Wrapf(ErrGrantExpression, "device %s: duplicate arg %q", name, key)
		}
		if err := setDeviceField(&d, k, val); err != nil {
			return device{}, err
		}
		seen[k] = true
	}

	if !seen[deviceArgPath] {
		return device{}, crex.Wrapf(ErrGrantExpression, "device %s: path required", name)
	}
	if !seen[deviceArgType] {
		return device{}, crex.Wrapf(ErrGrantExpression, "device %s: type required", name)
	}
	if !seen[deviceArgMajor] {
		return device{}, crex.Wrapf(ErrGrantExpression, "device %s: major required", name)
	}
	if !seen[deviceArgMinor] {
		return device{}, crex.Wrapf(ErrGrantExpression, "device %s: minor required", name)
	}
	if !seen[deviceArgFileMode] {
		return device{}, crex.Wrapf(ErrGrantExpression, "device %s: file_mode required", name)
	}

	return d, nil
}

// Sets a single field on a [device] from a key-value pair.
func setDeviceField(d *device, key deviceArg, val string) error {
	switch key {
	case deviceArgPath:
		p, err := validateDeviceContainerPath(val)
		if err != nil {
			return err
		}
		d.Path = p
	case deviceArgType:
		t, err := parseDeviceNodeType(val)
		if err != nil {
			return err
		}
		d.Type = t
	case deviceArgMajor:
		return parseUint32(&d.Major, val)
	case deviceArgMinor:
		return parseUint32(&d.Minor, val)
	case deviceArgFileMode:
		if err := parseUint16(&d.FileMode, val); err != nil {
			return err
		}
		return validateFileMode(d.FileMode)
	case deviceArgUID:
		return parseUint32(&d.UID, val)
	case deviceArgGID:
		return parseUint32(&d.GID, val)
	default:
		return crex.Wrapf(ErrGrantExpression, "unknown device arg %q", key)
	}
	return nil
}

// Converts a string to a [deviceNodeType], returning an error for unknown values.
func parseDeviceNodeType(s string) (deviceNodeType, error) {
	t := deviceNodeType(s)
	switch t {
	case deviceNodeChar, deviceNodeBlock, deviceNodeUnbuffered, deviceNodeFIFO:
		return t, nil
	default:
		return "", crex.Wrapf(ErrGrantExpression, "unknown device type %q", s)
	}
}
