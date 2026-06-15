package units

import "testing"

func TestIsKnown(t *testing.T) {
	known := []string{"Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "k", "K", "M", "G", "T", "P", "E", "m", "u", "n"}
	for _, s := range known {
		if !IsKnown(s) {
			t.Errorf("IsKnown(%q) = false, want true", s)
		}
	}
	unknown := []string{"", "x", "Xi", "kk", "Mib"}
	for _, s := range unknown {
		if IsKnown(s) {
			t.Errorf("IsKnown(%q) = true, want false", s)
		}
	}
}

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
