package codec

import "testing"

func TestFormat_String(t *testing.T) {
	tests := []struct {
		f    Format
		want string
	}{
		{JSON, "json"},
		{YAML, "yaml"},
		{Format(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.f.String(); got != tt.want {
			t.Errorf("Format(%d).String() = %q, want %q", tt.f, got, tt.want)
		}
	}
}
