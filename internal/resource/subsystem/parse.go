package subsystem

import (
	"strconv"
	"strings"

	"github.com/cruciblehq/crex"
)

// Parses s as a base-0 uint64 and stores it in dst.
func parseUint64(dst *uint64, ctx string, s string) error {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 0, 64)
	if err != nil {
		return crex.Wrapf(ErrSandboxExpression, "%s: %v", ctx, err)
	}
	*dst = v
	return nil
}

// Parses s as a base-0 uint32 and stores it in dst.
func parseUint32(dst *uint32, ctx string, s string) error {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 0, 32)
	if err != nil {
		return crex.Wrapf(ErrSandboxExpression, "%s: %v", ctx, err)
	}
	*dst = uint32(v)
	return nil
}

// Parses s as a base-0 uint16 and stores it in dst.
func parseUint16(dst *uint16, ctx string, s string) error {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 0, 16)
	if err != nil {
		return crex.Wrapf(ErrSandboxExpression, "%s: %v", ctx, err)
	}
	*dst = uint16(v)
	return nil
}

// Parses s as a float64 and stores it in dst.
func parseFloat64(dst *float64, ctx string, s string) error {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return crex.Wrapf(ErrSandboxExpression, "%s: %v", ctx, err)
	}
	*dst = v
	return nil
}

// Parses s as a boolean ("true"/"1" or "false"/"0") and stores it in dst.
func parseBool(dst *bool, ctx string, s string) error {
	switch strings.TrimSpace(s) {
	case "true", "1":
		*dst = true
	case "false", "0":
		*dst = false
	default:
		return crex.Wrapf(ErrSandboxExpression, "%s: invalid boolean %q", ctx, s)
	}
	return nil
}

// Parses s as an int64 and stores it in dst.
func parseInt64(dst *int64, ctx string, s string) error {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 0, 64)
	if err != nil {
		return crex.Wrapf(ErrSandboxExpression, "%s: %v", ctx, err)
	}
	*dst = v
	return nil
}
