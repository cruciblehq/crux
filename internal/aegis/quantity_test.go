package aegis

import "testing"

func TestMultiplier(t *testing.T) {
	cases := []struct {
		suffix QuantitySuffix
		want   uint64
		ok     bool
	}{
		// IEC binary multipliers.
		{SuffixKi, 1 << 10, true},
		{SuffixMi, 1 << 20, true},
		{SuffixGi, 1 << 30, true},
		{SuffixTi, 1 << 40, true},
		{SuffixPi, 1 << 50, true},
		{SuffixEi, 1 << 60, true},

		// SI decimal multipliers.
		{SuffixKLower, 1_000, true},
		{SuffixK, 1_000, true},
		{SuffixM, 1_000_000, true},
		{SuffixG, 1_000_000_000, true},
		{SuffixT, 1_000_000_000_000, true},
		{SuffixP, 1_000_000_000_000_000, true},
		{SuffixE, 1_000_000_000_000_000_000, true},

		// Sub-unit suffixes — no integer multiplier.
		{SuffixMilli, 0, false},
		{SuffixMicro, 0, false},
		{SuffixNano, 0, false},

		// Unknown suffix.
		{"foo", 0, false},
		{"", 0, false},
	}
	for _, tc := range cases {
		got, ok := tc.suffix.Multiplier()
		if ok != tc.ok {
			t.Errorf("Multiplier(%q): ok = %v, want %v", tc.suffix, ok, tc.ok)
			continue
		}
		if tc.ok && got != tc.want {
			t.Errorf("Multiplier(%q): got %d, want %d", tc.suffix, got, tc.want)
		}
	}
}
