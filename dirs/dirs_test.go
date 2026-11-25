package dirs

import (
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

func TestDirs(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Local", Local(), filepath.Join(xdg.DataHome, "crux", "local")},
		{"Lima", Lima(), filepath.Join(xdg.DataHome, "crux", "lima")},
		{"VM", VM(), filepath.Join(xdg.DataHome, "crux", "vm")},
		{"Registry", Registry(), filepath.Join(xdg.CacheHome, "crux", "registry")},
		{"Temp", Temp(), filepath.Join(xdg.CacheHome, "crux", "tmp")},
		{"Trust", Trust(), filepath.Join(xdg.DataHome, "crux", "trust")},
		{"Machines", Machines(), filepath.Join(xdg.DataHome, "crux", "machines")},
		{"Secrets", Secrets(), filepath.Join(xdg.DataHome, "crux", "secrets")},
		{"Config", Config(), filepath.Join(xdg.ConfigHome, "crux")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s() = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}
