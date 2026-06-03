package compute

import (
	"reflect"
	"testing"
)

func TestMergeEnv(t *testing.T) {
	tests := []struct {
		name      string
		base      []string
		overrides []string
		want      []string
	}{
		{
			name:      "empty base and overrides",
			base:      nil,
			overrides: nil,
			want:      []string{},
		},
		{
			name:      "overrides only",
			base:      nil,
			overrides: []string{"A=1"},
			want:      []string{"A=1"},
		},
		{
			name:      "base only",
			base:      []string{"A=1", "B=2"},
			overrides: nil,
			want:      []string{"A=1", "B=2"},
		},
		{
			name:      "no overlapping keys",
			base:      []string{"A=1", "B=2"},
			overrides: []string{"C=3"},
			want:      []string{"A=1", "B=2", "C=3"},
		},
		{
			name:      "override wins on matching key",
			base:      []string{"A=1", "B=2"},
			overrides: []string{"A=99"},
			want:      []string{"B=2", "A=99"},
		},
		{
			name:      "all keys overridden",
			base:      []string{"A=1", "B=2"},
			overrides: []string{"B=20", "A=10"},
			want:      []string{"B=20", "A=10"},
		},
		{
			name:      "base entry without = is kept",
			base:      []string{"NOEQUALS", "A=1"},
			overrides: []string{"A=2"},
			want:      []string{"NOEQUALS", "A=2"},
		},
		{
			name:      "value with = in it is handled correctly",
			base:      []string{"A=x=y"},
			overrides: []string{"A=z"},
			want:      []string{"A=z"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeEnv(tc.base, tc.overrides)
			if got == nil {
				got = []string{}
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("MergeEnv(%v, %v) = %v; want %v", tc.base, tc.overrides, got, tc.want)
			}
		})
	}
}
