package subsystem

import (
	"strconv"
	"strings"

	"github.com/cruciblehq/crex"
)

// Parses s as a base-0 uint64 and stores it in dst.
func parseUint64(dst *uint64, s string) error {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 0, 64)
	if err != nil {
		return crex.Wrapf(ErrGrantExpression, "%v", err)
	}
	*dst = v
	return nil
}

// Parses s as a base-0 uint32 and stores it in dst.
func parseUint32(dst *uint32, s string) error {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 0, 32)
	if err != nil {
		return crex.Wrapf(ErrGrantExpression, "%v", err)
	}
	*dst = uint32(v)
	return nil
}

// Parses s as a base-0 uint16 and stores it in dst.
func parseUint16(dst *uint16, s string) error {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 0, 16)
	if err != nil {
		return crex.Wrapf(ErrGrantExpression, "%v", err)
	}
	*dst = uint16(v)
	return nil
}

// Parses s as a float64 and stores it in dst.
func parseFloat64(dst *float64, s string) error {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return crex.Wrapf(ErrGrantExpression, "%v", err)
	}
	*dst = v
	return nil
}

// Parses s as a boolean ("true"/"1" or "false"/"0") and stores it in dst.
func parseBool(dst *bool, s string) error {
	switch strings.TrimSpace(s) {
	case "true", "1":
		*dst = true
	case "false", "0":
		*dst = false
	default:
		return crex.Wrapf(ErrGrantExpression, "invalid boolean %q", s)
	}
	return nil
}
