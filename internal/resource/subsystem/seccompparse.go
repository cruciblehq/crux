package subsystem

import (
	"strconv"
	"strings"

	"github.com/cruciblehq/crex"
)

// Resolves a compact seccomp expression into expanded rules.
//
// Parses the syscall name from expr, then parses each arg filter. Bracket
// lists are expanded into the cartesian product. Returns one [seccomp] per
// combination. With no args, returns a single unconditional rule.
func parseSeccomp(expr string, args []string) ([]seccomp, error) {
	var r seccomp
	if err := r.UnmarshalText([]byte(expr)); err != nil {
		return nil, err
	}
	if args == nil {
		return []seccomp{r}, nil
	}

	argSets := make([][]seccompArg, len(args))
	for i, arg := range args {
		alts, err := parseSeccompArgExpr(arg)
		if err != nil {
			return nil, crex.Wrapf(ErrSeccompExpression, "syscall %s: %w", r.Syscall, err)
		}
		argSets[i] = alts
	}

	var rules []seccomp
	expandSeccomp(&rules, r.Syscall, argSets, nil, 0)
	return rules, nil
}

// Parses one argument filter expression into one or more alternatives.
//
// The list form expands to one seccompArg per value, all sharing the same
// arg position and operator. Each alternative produces a separate [seccomp]
// rule in the final BPF filter, OR'd together at load time.
func parseSeccompArgExpr(s string) ([]seccompArg, error) {
	if i := strings.IndexByte(s, '['); i >= 0 {
		return parseSeccompArgList(s, i)
	}
	a, err := parseSeccompArgScalar(s)
	if err != nil {
		return nil, err
	}
	return []seccompArg{a}, nil
}

// Parses the shared "<arg> <op>" header from an argument filter expression.
// Returns the validated arg index, parsed operator, and nil error on success.
func parseArgHeader(posStr, opStr string) (uint8, seccompOp, error) {
	argIdx, err := strconv.ParseUint(posStr, 0, 8)
	if err != nil {
		return 0, "", crex.Wrapf(ErrSeccompArgFilter, "bad arg position %q: %w", posStr, err)
	}
	if argIdx > 5 {
		return 0, "", crex.Wrapf(ErrSeccompArgFilter, "arg position %d out of range (0-5)", argIdx)
	}
	op, err := parseSeccompOp(opStr)
	if err != nil {
		return 0, "", err
	}
	return uint8(argIdx), op, nil
}

// Parses a scalar argument expression: "<arg> <op> <val> [<mask>]".
func parseSeccompArgScalar(s string) (seccompArg, error) {
	fields := strings.Fields(s)

	if len(fields) < 3 {
		return seccompArg{}, crex.Wrapf(ErrSeccompArgFilter, "not enough fields in %q", s)
	}
	if len(fields) > 4 {
		return seccompArg{}, crex.Wrapf(ErrSeccompArgFilter, "too many fields in %q", s)
	}

	argIdx, op, err := parseArgHeader(fields[0], fields[1])
	if err != nil {
		return seccompArg{}, err
	}

	val, err := strconv.ParseUint(fields[2], 0, 64)
	if err != nil {
		return seccompArg{}, crex.Wrapf(ErrSeccompArgFilter, "bad val %q: %w", fields[2], err)
	}

	a := seccompArg{
		Arg: argIdx,
		Op:  op,
		Val: val,
	}

	if len(fields) == 4 {
		if op != seccompOpMaskedEq {
			return seccompArg{}, crex.Wrapf(ErrSeccompArgFilter, "mask used with %q", op)
		}
		mask, err := strconv.ParseUint(fields[3], 0, 64)
		if err != nil {
			return seccompArg{}, crex.Wrapf(ErrSeccompArgFilter, "bad mask %q: %w", fields[3], err)
		}
		a.Mask = mask
	}

	return a, nil
}

// Parses a list argument expression: "<arg> <op> [<val>, <val>, ...]".
//
// Returns one seccompArg per value in the bracket list, all sharing the
// same arg position and operator. An empty or unclosed list is an error.
// The masked_eq operator is rejected because the list syntax has no way
// to specify a per-value mask.
func parseSeccompArgList(s string, bracketStart int) ([]seccompArg, error) {
	prefix := strings.TrimSpace(s[:bracketStart])
	fields := strings.Fields(prefix)
	if len(fields) != 2 {
		return nil, crex.Wrapf(ErrSeccompArgFilter, "invalid list prefix %q", prefix)
	}

	argIdx, op, err := parseArgHeader(fields[0], fields[1])
	if err != nil {
		return nil, err
	}
	if op == seccompOpMaskedEq {
		return nil, crex.Wrapf(ErrSeccompArgFilter, "masked_eq cannot be used with list syntax")
	}

	bracketEnd := strings.LastIndexByte(s, ']')
	if bracketEnd < bracketStart {
		return nil, crex.Wrapf(ErrSeccompArgFilter, "unclosed bracket in %q", s)
	}

	inner := s[bracketStart+1 : bracketEnd]
	parts := strings.Split(inner, ",")
	result := make([]seccompArg, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		val, err := strconv.ParseUint(p, 0, 64)
		if err != nil {
			return nil, crex.Wrapf(ErrSeccompArgFilter, "bad val %q: %w", p, err)
		}
		result = append(result, seccompArg{Arg: argIdx, Op: op, Val: val})
	}
	if len(result) == 0 {
		return nil, crex.Wrapf(ErrSeccompArgFilter, "empty list in %q", s)
	}
	return result, nil
}

// Expands the cartesian product of arg alternatives into [seccomp] rules.
//
// Each combination of one alternative per sub-arg position becomes a separate
// rule. Within each rule, all arg conditions are AND'd. Across rules for the
// same syscall, the BPF compiler OR's them.
func expandSeccomp(rules *[]seccomp, syscall string, argSets [][]seccompArg, current []seccompArg, depth int) {
	if depth == len(argSets) {
		rule := seccomp{Syscall: syscall, Args: make([]seccompArg, len(current))}
		copy(rule.Args, current)
		*rules = append(*rules, rule)
		return
	}
	for _, alt := range argSets[depth] {
		expandSeccomp(rules, syscall, argSets, append(current, alt), depth+1)
	}
}
