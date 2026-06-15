package units

// Unit suffix on a quantity literal.
//
// AGL recognises a set of suffixes that follow a decimal integer to form a
// single TokenQuantity (e.g. "1Gi", "500m"). The underlying string is the
// canonical source spelling, so a QuantitySuffix can be formatted directly
// with %s. Each subsystem decides which suffixes a given knob accepts and
// what each suffix multiplies by (bytes, millicores, seconds, ...); the
// lexer attaches no semantics.
type QuantitySuffix string

const (
	SuffixKi QuantitySuffix = "Ki" // IEC power-of-1024 multiplier (2^10).
	SuffixMi QuantitySuffix = "Mi" // IEC power-of-1024 multiplier (2^20).
	SuffixGi QuantitySuffix = "Gi" // IEC power-of-1024 multiplier (2^30).
	SuffixTi QuantitySuffix = "Ti" // IEC power-of-1024 multiplier (2^40).
	SuffixPi QuantitySuffix = "Pi" // IEC power-of-1024 multiplier (2^50).
	SuffixEi QuantitySuffix = "Ei" // IEC power-of-1024 multiplier (2^60).
)

const (
	SuffixKLower QuantitySuffix = "k" // SI decimal multiplier (10^3).
	SuffixK      QuantitySuffix = "K" // SI decimal multiplier (10^3).
	SuffixM      QuantitySuffix = "M" // SI decimal multiplier (10^6).
	SuffixG      QuantitySuffix = "G" // SI decimal multiplier (10^9).
	SuffixT      QuantitySuffix = "T" // SI decimal multiplier (10^12).
	SuffixP      QuantitySuffix = "P" // SI decimal multiplier (10^15).
	SuffixE      QuantitySuffix = "E" // SI decimal multiplier (10^18).
)

const (
	SuffixMilli QuantitySuffix = "m" // SI decimal sub-unit multiplier (10^-3).
	SuffixMicro QuantitySuffix = "u" // SI decimal sub-unit multiplier (10^-6).
	SuffixNano  QuantitySuffix = "n" // SI decimal sub-unit multiplier (10^-9).
)

// Set of suffixes the lexer accepts.
//
// Admits any of these after a decimal integer; subsystems decide which
// knobs accept which suffixes.
var knownSuffixes = map[QuantitySuffix]struct{}{
	SuffixKi:     {},
	SuffixMi:     {},
	SuffixGi:     {},
	SuffixTi:     {},
	SuffixPi:     {},
	SuffixEi:     {},
	SuffixKLower: {},
	SuffixK:      {},
	SuffixM:      {},
	SuffixG:      {},
	SuffixT:      {},
	SuffixP:      {},
	SuffixE:      {},
	SuffixMilli:  {},
	SuffixMicro:  {},
	SuffixNano:   {},
}

// Whether s is a recognised quantity suffix.
func IsKnown(s string) bool {
	_, ok := knownSuffixes[QuantitySuffix(s)]
	return ok
}

// Returns the integer scale factor for the suffix.
//
// Returns the factor and true for IEC binary suffixes (Ki–Ei, powers of 1024)
// and SI decimal suffixes (k/K–E, powers of 1000). Returns zero and false for
// sub-unit suffixes (m, u, n) and for any value not in the recognised set;
// sub-unit suffixes represent fractions of the base unit and cannot be
// expressed as an integer multiplier.
func (s QuantitySuffix) Multiplier() (uint64, bool) {
	switch s {
	case SuffixKi:
		return 1 << 10, true
	case SuffixMi:
		return 1 << 20, true
	case SuffixGi:
		return 1 << 30, true
	case SuffixTi:
		return 1 << 40, true
	case SuffixPi:
		return 1 << 50, true
	case SuffixEi:
		return 1 << 60, true
	case SuffixKLower, SuffixK:
		return 1_000, true
	case SuffixM:
		return 1_000_000, true
	case SuffixG:
		return 1_000_000_000, true
	case SuffixT:
		return 1_000_000_000_000, true
	case SuffixP:
		return 1_000_000_000_000_000, true
	case SuffixE:
		return 1_000_000_000_000_000_000, true
	default:
		return 0, false
	}
}
