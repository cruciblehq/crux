package subsystem

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/crux/internal/manifest"
)

// Matches a valid syscall name: lowercase letters, digits, and underscores.
var validSyscallName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Comparison operator for seccomp argument filters.
type SeccompOp string

const (
	SeccompOpEqual    SeccompOp = "eq"        // Exact match.
	SeccompOpNotEqual SeccompOp = "ne"        // Not equal.
	SeccompOpGreater  SeccompOp = "gt"        // Greater than.
	SeccompOpGreaterE SeccompOp = "ge"        // Greater than or equal.
	SeccompOpLess     SeccompOp = "lt"        // Less than.
	SeccompOpLessE    SeccompOp = "le"        // Less than or equal.
	SeccompOpMaskedEq SeccompOp = "masked_eq" // (arg & Mask) == Val.
)

// Allows one syscall, optionally constrained by argument filters.
//
// When Args is empty, every invocation of the named syscall is permitted.
// When Args is non-empty, all conditions must match (AND) for the syscall
// to be allowed; invocations that don't match the conditions are killed.
// Multiple rules for the same syscall with different Args are OR'd by the
// BPF compiler: the syscall is allowed if any rule matches.
type Seccomp struct {
	Syscall string       `codec:"syscall"`        // Syscall name (e.g., "socket", "read", "openat").
	Args    []SeccompArg `codec:"args,omitempty"` // Argument conditions. All must match (AND).
}

// Constrains one syscall argument by position.
//
// Arg selects which argument (0–5). The comparison is Op(arg, Val). For
// SeccompOpMaskedEq the comparison is (arg & Mask) == Val.
type SeccompArg struct {
	Arg  uint8     `codec:"arg"`            // Argument position (0-5).
	Op   SeccompOp `codec:"op"`             // Comparison operator.
	Val  uint64    `codec:"val"`            // Comparison value. For MaskedEq, the expected result.
	Mask uint64    `codec:"mask,omitempty"` // Bitmask. Only used with MaskedEq: (arg & Mask) == Val.
}

// Parses a compact expression into a Seccomp rule.
//
// The string form is a bare syscall name (e.g., "read", "socket"). Argument
// filters are specified via the structured sub-args form only.
func (r *Seccomp) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		return crex.Wrapf(ErrSeccompExpression, "expression is empty")
	}
	if !validSyscallName.MatchString(s) {
		return crex.Wrapf(ErrSeccompExpression, "invalid syscall name %q", s)
	}
	r.Syscall = s
	r.Args = nil
	return nil
}

// Implements [Subsystem] for [manifest.DomainSeccomp].
type SeccompSubsystem struct{}

// Builds seccomp grants by parsing and expanding the input expression.
//
// Validates the syscall name and argument filters. Bracket lists in args
// are expanded into the cartesian product, producing one grant per
// combination. Each returned grant has normalized arg strings.
func (s *SeccompSubsystem) Build(_ context.Context, domain Domain, input manifest.Grant) ([]manifest.Grant, error) {
	crex.Assertf(domain == DomainSeccomp, "seccomp: unexpected domain %q", domain)
	rules, err := parseSeccomp(input.Expr, input.Args)
	if err != nil {
		return nil, err
	}
	grants := make([]manifest.Grant, len(rules))
	for i, r := range rules {
		grants[i] = manifest.Grant{
			Subsystem: string(domain),
			Expr:      r.Syscall,
			Args:      serializeSeccompArgs(r.Args),
		}
	}
	return grants, nil
}

// Serializes SeccompArgs back to their text form.
//
// Each arg becomes "<position> <op> <val>" or "<position> <op> <val> <mask>"
// for masked_eq. This is the inverse of parseSeccompArgScalar.
func serializeSeccompArgs(args []SeccompArg) []string {
	if len(args) == 0 {
		return nil
	}
	out := make([]string, len(args))
	for i, a := range args {
		s := strconv.FormatUint(uint64(a.Arg), 10) + " " + string(a.Op) + " " + strconv.FormatUint(a.Val, 10)
		if a.Op == SeccompOpMaskedEq {
			s += " " + strconv.FormatUint(a.Mask, 10)
		}
		out[i] = s
	}
	return out
}

// Applies seccomp grants to a runtime.
//
// Not implemented yet; for now, seccomp grants are built but not applied.
func (s *SeccompSubsystem) Apply(_ context.Context, _ Domain, _ manifest.Grant) error {
	crex.Assert(false, "not implemented")
	return nil
}

// Resolves a compact seccomp expression into expanded rules.
//
// Parses the syscall name from expr, then parses each arg filter. Bracket
// lists are expanded into the cartesian product. Returns one [Seccomp] per
// combination. With no args, returns a single unconditional rule.
func parseSeccomp(expr string, args []string) ([]Seccomp, error) {
	var r Seccomp
	if err := r.UnmarshalText([]byte(expr)); err != nil {
		return nil, err
	}
	if args == nil {
		return []Seccomp{r}, nil
	}

	argSets := make([][]SeccompArg, len(args))
	for i, arg := range args {
		alts, err := parseSeccompArgExpr(arg)
		if err != nil {
			return nil, crex.Wrapf(ErrSeccompExpression, "syscall %s: %w", r.Syscall, err)
		}
		argSets[i] = alts
	}

	var rules []Seccomp
	expandSeccomp(&rules, r.Syscall, argSets, nil, 0)
	return rules, nil
}

// Parses one argument filter expression into one or more alternatives.
//
// The list form expands to one SeccompArg per value, all sharing the same
// arg position and operator. Each alternative produces a separate
// [Seccomp] rule in the final BPF filter, OR'd together at load time.
func parseSeccompArgExpr(s string) ([]SeccompArg, error) {
	if i := strings.IndexByte(s, '['); i >= 0 {
		return parseSeccompArgList(s, i)
	}
	a, err := parseSeccompArgScalar(s)
	if err != nil {
		return nil, err
	}
	return []SeccompArg{a}, nil
}

// Parses a scalar argument expression: "<arg> <op> <val> [<mask>]".
func parseSeccompArgScalar(s string) (SeccompArg, error) {
	fields := strings.Fields(s)

	if len(fields) < 3 {
		return SeccompArg{}, crex.Wrapf(ErrSeccompArgFilter, "not enough fields in %q", s)
	}
	if len(fields) > 4 {
		return SeccompArg{}, crex.Wrapf(ErrSeccompArgFilter, "too many fields in %q", s)
	}

	argIdx, err := strconv.ParseUint(fields[0], 0, 8)
	if err != nil {
		return SeccompArg{}, crex.Wrapf(ErrSeccompArgFilter, "bad arg position %q: %w", fields[0], err)
	}
	if argIdx > 5 {
		return SeccompArg{}, crex.Wrapf(ErrSeccompArgFilter, "arg position %d out of range (0-5)", argIdx)
	}

	op, err := parseSeccompOp(fields[1])
	if err != nil {
		return SeccompArg{}, err
	}

	val, err := strconv.ParseUint(fields[2], 0, 64)
	if err != nil {
		return SeccompArg{}, crex.Wrapf(ErrSeccompArgFilter, "bad val %q: %w", fields[2], err)
	}

	a := SeccompArg{
		Arg: uint8(argIdx),
		Op:  op,
		Val: val,
	}

	if len(fields) == 4 {
		if op != SeccompOpMaskedEq {
			return SeccompArg{}, crex.Wrapf(ErrSeccompArgFilter, "mask used with %q", op)
		}
		mask, err := strconv.ParseUint(fields[3], 0, 64)
		if err != nil {
			return SeccompArg{}, crex.Wrapf(ErrSeccompArgFilter, "bad mask %q: %w", fields[3], err)
		}
		a.Mask = mask
	}

	return a, nil
}

// Parses a list argument expression: "<arg> <op> [<val>, <val>, ...]".
//
// Returns one SeccompArg per value in the bracket list, all sharing the
// same arg position and operator. An empty or unclosed list is an error.
func parseSeccompArgList(s string, bracketStart int) ([]SeccompArg, error) {
	prefix := strings.TrimSpace(s[:bracketStart])
	fields := strings.Fields(prefix)
	if len(fields) != 2 {
		return nil, crex.Wrapf(ErrSeccompArgFilter, "invalid list prefix %q", prefix)
	}

	argIdx, err := strconv.ParseUint(fields[0], 0, 8)
	if err != nil {
		return nil, crex.Wrapf(ErrSeccompArgFilter, "bad arg position %q: %w", fields[0], err)
	}
	if argIdx > 5 {
		return nil, crex.Wrapf(ErrSeccompArgFilter, "arg position %d out of range (0-5)", argIdx)
	}

	op, err := parseSeccompOp(fields[1])
	if err != nil {
		return nil, err
	}

	bracketEnd := strings.LastIndexByte(s, ']')
	if bracketEnd < bracketStart {
		return nil, crex.Wrapf(ErrSeccompArgFilter, "unclosed bracket in %q", s)
	}

	inner := s[bracketStart+1 : bracketEnd]
	parts := strings.Split(inner, ",")
	result := make([]SeccompArg, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		val, err := strconv.ParseUint(p, 0, 64)
		if err != nil {
			return nil, crex.Wrapf(ErrSeccompArgFilter, "bad val %q: %w", p, err)
		}
		result = append(result, SeccompArg{Arg: uint8(argIdx), Op: op, Val: val})
	}
	if len(result) == 0 {
		return nil, crex.Wrapf(ErrSeccompArgFilter, "empty list in %q", s)
	}
	return result, nil
}

// Converts a string to a SeccompOp, returning an error for unknown operators.
func parseSeccompOp(s string) (SeccompOp, error) {
	op := SeccompOp(s)
	switch op {
	case SeccompOpEqual, SeccompOpNotEqual,
		SeccompOpGreater, SeccompOpGreaterE,
		SeccompOpLess, SeccompOpLessE,
		SeccompOpMaskedEq:
		return op, nil
	default:
		return "", crex.Wrapf(ErrSeccompArgFilter, "unknown op %q", s)
	}
}

// Expands the cartesian product of arg alternatives into [Seccomp] rules.
//
// Each combination of one alternative per sub-arg position becomes a
// separate rule. Within each rule, all arg conditions are AND'd. Across
// rules for the same syscall, the BPF compiler OR's them.
func expandSeccomp(rules *[]Seccomp, syscall string, argSets [][]SeccompArg, current []SeccompArg, depth int) {
	if depth == len(argSets) {
		rule := Seccomp{Syscall: syscall, Args: make([]SeccompArg, len(current))}
		copy(rule.Args, current)
		*rules = append(*rules, rule)
		return
	}
	for _, alt := range argSets[depth] {
		expandSeccomp(rules, syscall, argSets, append(current, alt), depth+1)
	}
}
