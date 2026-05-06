package aegis

// Unit suffix on a quantity literal.
//
// The lexer recognises a closed set of suffixes that follow a decimal integer
// to form a single TokenQuantity (e.g. "1Gi", "500m"). The underlying string
// is the canonical source spelling, so a QuantitySuffix can be formatted
// directly with %s. Each subsystem decides which suffixes a given knob
// accepts and what each suffix multiplies by (bytes, millicores, seconds,
// ...); the lexer attaches no semantics.
type QuantitySuffix string

// IEC binary multipliers. Conventionally power-of-1024 byte counts.
const (
	SuffixKi QuantitySuffix = "Ki" // 2^10
	SuffixMi QuantitySuffix = "Mi" // 2^20
	SuffixGi QuantitySuffix = "Gi" // 2^30
	SuffixTi QuantitySuffix = "Ti" // 2^40
	SuffixPi QuantitySuffix = "Pi" // 2^50
	SuffixEi QuantitySuffix = "Ei" // 2^60
)

// SI decimal multipliers. Conventionally power-of-1000 counts. Both the
// lower-case and upper-case forms of "kilo" are accepted.
const (
	SuffixKLower QuantitySuffix = "k" // 10^3
	SuffixK      QuantitySuffix = "K" // 10^3
	SuffixM      QuantitySuffix = "M" // 10^6
	SuffixG      QuantitySuffix = "G" // 10^9
	SuffixT      QuantitySuffix = "T" // 10^12
	SuffixP      QuantitySuffix = "P" // 10^15
	SuffixE      QuantitySuffix = "E" // 10^18
)

// SI sub-unit multipliers. Conventionally fractions of the base unit.
const (
	SuffixMilli QuantitySuffix = "m" // 10^-3
	SuffixMicro QuantitySuffix = "u" // 10^-6
	SuffixNano  QuantitySuffix = "n" // 10^-9
)

// Set of suffixes the lexer accepts.
//
// The lexer admits any of these after a decimal integer; subsystems decide
// which knobs accept which suffixes.
var quantitySuffixes = map[QuantitySuffix]struct{}{
	SuffixKi: {}, SuffixMi: {}, SuffixGi: {}, SuffixTi: {}, SuffixPi: {}, SuffixEi: {},
	SuffixKLower: {}, SuffixK: {}, SuffixM: {}, SuffixG: {}, SuffixT: {}, SuffixP: {}, SuffixE: {},
	SuffixMilli: {}, SuffixMicro: {}, SuffixNano: {},
}

// Whether s is a recognised quantity suffix.
func isQuantitySuffix(s string) bool {
	_, ok := quantitySuffixes[QuantitySuffix(s)]
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
