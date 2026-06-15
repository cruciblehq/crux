package cgroup

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/cruciblehq/crux/crex"
)

// Regex to match major and minor positional operands.
//
// Matches "major minor" optionally followed by the rest of the value.
// Major and minor are unsigned integers. For example, "8 0 rbps=1000"
// would specify major 8 and minor 0 with optional args rbps=1000.
var reMajorMinor = regexp.MustCompile(`^(\d+)\s+(\d+)(?:\s+(.+))?$`)

// Parses required positional major and minor at the head of value.
//
// Returns the parsed major and minor and the remaining trimmed content. The
// remainder is the empty string when value carries only the pair, and the
// trailing arg string otherwise. Returns ErrInvalidGrant when the pair is
// missing or its components do not fit in uint32.
func parseMajorMinor(value string) (uint32, uint32, string, error) {
	m := reMajorMinor.FindStringSubmatch(strings.TrimSpace(value))
	if m == nil {
		return 0, 0, "", crex.Wrapf(ErrInvalidGrant, "expected major minor positional operands")
	}
	maj, err := strconv.ParseUint(m[1], 10, 32)
	if err != nil {
		return 0, 0, "", crex.Wrap(ErrInvalidGrant, err)
	}
	min, err := strconv.ParseUint(m[2], 10, 32)
	if err != nil {
		return 0, 0, "", crex.Wrap(ErrInvalidGrant, err)
	}
	return uint32(maj), uint32(min), strings.TrimSpace(m[3]), nil
}

// Parses space-separated key=value tokens by dispatching each to a handler.
//
// fields maps each accepted key to a handler responsible for parsing and
// storing the value. Tokens must be non-empty key=value pairs. Duplicate
// keys and keys absent from fields are rejected with ErrInvalidGrant.
// Handler errors are propagated unchanged so per-knob parsers control their
// own wrapping. An empty value is a successful no-op.
func parseArgs(value string, fields map[string]func(string) error) error {
	tokens := strings.Fields(value)
	seen := make(map[string]struct{}, len(tokens))
	for _, tok := range tokens {
		key, val, ok := strings.Cut(tok, "=")
		if !ok || key == "" || val == "" {
			return crex.Wrapf(ErrInvalidGrant, "invalid optional arg %q, expected key=value", tok)
		}
		if _, ok := seen[key]; ok {
			return crex.Wrapf(ErrInvalidGrant, "duplicate key %q", key)
		}
		seen[key] = struct{}{}
		parseCallback, ok := fields[key]
		if !ok {
			return crex.Wrapf(ErrInvalidGrant, "unknown key %q", key)
		}
		if err := parseCallback(val); err != nil {
			return err
		}
	}
	return nil
}

// Parses an unsigned 64-bit integer into dst.
//
// Uses strconv base 0 so decimal, 0x, 0o, and 0b prefixes are all accepted.
// Wraps strconv errors with ErrInvalidGrant.
func parseUint64(dst *uint64, s string) error {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 0, 64)
	if err != nil {
		return crex.Wrap(ErrInvalidGrant, err)
	}
	*dst = v
	return nil
}

// Parses an unsigned 32-bit integer into dst.
//
// Uses strconv base 0 so decimal, 0x, 0o, and 0b prefixes are all accepted.
// Wraps strconv errors with ErrInvalidGrant.
func parseUint32(dst *uint32, s string) error {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 0, 32)
	if err != nil {
		return crex.Wrap(ErrInvalidGrant, err)
	}
	*dst = uint32(v)
	return nil
}

// Parses an unsigned 16-bit integer into dst.
//
// Uses strconv base 0 so decimal, 0x, 0o, and 0b prefixes are all accepted.
// Wraps strconv errors with ErrInvalidGrant.
func parseUint16(dst *uint16, s string) error {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 0, 16)
	if err != nil {
		return crex.Wrap(ErrInvalidGrant, err)
	}
	*dst = uint16(v)
	return nil
}

// Parses a 64-bit float into dst, wrapping strconv errors with ErrInvalidGrant.
func parseFloat64(dst *float64, s string) error {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return crex.Wrap(ErrInvalidGrant, err)
	}
	*dst = v
	return nil
}

// Parses a boolean value from a string, storing the result in dst.
//
// Parses case-sensitive "true"/"1" as true and "false"/"0" as false. Returns
// an error for any other value.
func parseBool(dst *bool, s string) error {
	switch strings.TrimSpace(s) {
	case "true", "1":
		*dst = true
	case "false", "0":
		*dst = false
	default:
		return crex.Wrapf(ErrInvalidGrant, "invalid boolean %q", s)
	}
	return nil
}
