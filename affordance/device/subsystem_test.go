package device

import (
	"errors"
	"os"
	"testing"

	"github.com/cruciblehq/crux/affordance/agl"
	"github.com/cruciblehq/crux/affordance/subsystem"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func parseGrant(t *testing.T, src string) *agl.Model {
	t.Helper()
	g, err := agl.Parse(src)
	if err != nil {
		t.Fatalf("agl.Parse(%q): %v", src, err)
	}
	return g
}

func newSub() (*Subsystem, *[]specs.LinuxDevice) {
	d := make([]specs.LinuxDevice, 0)
	return New(&d), &d
}

func TestBuildCharDevice(t *testing.T) {
	sub, devices := newSub()
	if err := sub.Build(parseGrant(t, ".device c /dev/nvidia0 195 0")); err != nil {
		t.Fatal(err)
	}
	if len(*devices) != 1 {
		t.Fatalf("want 1 device, got %d", len(*devices))
	}
	d := (*devices)[0]
	if d.Type != "c" || d.Path != "/dev/nvidia0" || d.Major != 195 || d.Minor != 0 {
		t.Errorf("unexpected device: %+v", d)
	}
	if d.FileMode != nil || d.UID != nil || d.GID != nil {
		t.Errorf("expected nil optional fields, got %+v", d)
	}
}

func TestBuildBlockDevice(t *testing.T) {
	sub, devices := newSub()
	if err := sub.Build(parseGrant(t, ".device b /dev/loop0 7 0")); err != nil {
		t.Fatal(err)
	}
	d := (*devices)[0]
	if d.Type != "b" || d.Path != "/dev/loop0" {
		t.Errorf("unexpected device: %+v", d)
	}
}

func TestBuildWithModeUIDGID(t *testing.T) {
	sub, devices := newSub()
	if err := sub.Build(parseGrant(t, ".device c /dev/fuse 10 229 mode=0660 uid=0 gid=44")); err != nil {
		t.Fatal(err)
	}
	d := (*devices)[0]
	if d.FileMode == nil || *d.FileMode != os.FileMode(0o660) {
		t.Errorf("FileMode: got %v, want 0660", d.FileMode)
	}
	if d.UID == nil || *d.UID != 0 {
		t.Errorf("UID: got %v, want 0", d.UID)
	}
	if d.GID == nil || *d.GID != 44 {
		t.Errorf("GID: got %v, want 44", d.GID)
	}
}

func TestBuildFIFO(t *testing.T) {
	sub, devices := newSub()
	if err := sub.Build(parseGrant(t, ".device p /dev/initctl 0 0")); err != nil {
		t.Fatal(err)
	}
	d := (*devices)[0]
	if d.Type != "p" || d.Path != "/dev/initctl" {
		t.Errorf("unexpected device: %+v", d)
	}
}

func TestKeyIsPath(t *testing.T) {
	sub, _ := newSub()
	g := parseGrant(t, ".device c /dev/nvidia0 195 0")
	if got := sub.Key(g); got != "/dev/nvidia0" {
		t.Fatalf("Key() = %q, want %q", got, "/dev/nvidia0")
	}
}

func TestName(t *testing.T) {
	sub, _ := newSub()
	if got := sub.Name(); got != subsystem.NameDevice {
		t.Fatalf("Name() = %q, want %q", got, subsystem.NameDevice)
	}
}

func TestBuildRejectsUnknownType(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(parseGrant(t, ".device x /dev/nvidia0 195 0"))
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsRelativePath(t *testing.T) {
	sub, _ := newSub()
	err := sub.Build(parseGrant(t, `.device c "dev/nvidia0" 195 0`))
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsWhereClause(t *testing.T) {
	sub, _ := newSub()
	g := parseGrant(t, ".device c /dev/nvidia0 195 0")
	g.Where = &agl.CompareExpr{}
	err := sub.Build(g)
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("err = %v, want ErrInvalidGrant", err)
	}
}

func TestBuildRejectsInvalidGrants(t *testing.T) {
	cases := []struct {
		name  string
		grant string
	}{
		{"type not a name", ".device 5 /dev/nvidia0 195 0"},
		{"path not a node path", ".device c 5 195 0"},
		{"path not clean", `.device c "/dev/../etc/x" 195 0`},
		{"path trailing slash", `.device c "/dev/" 195 0`},
		{"too few arguments", ".device c /dev/nvidia0 195"},
		{"too many arguments", ".device c /dev/nvidia0 195 0 5"},
		{"major not an integer", ".device c /dev/nvidia0 major 0"},
		{"major overflows int64", ".device c /dev/nvidia0 99999999999999999999 0"},
		{"minor not an integer", ".device c /dev/nvidia0 195 minor"},
		{"minor overflows int64", ".device c /dev/nvidia0 195 99999999999999999999"},
		{"mode not an integer", ".device c /dev/nvidia0 195 0 mode=rwx"},
		{"mode not octal", ".device c /dev/nvidia0 195 0 mode=8"},
		{"uid not an integer", ".device c /dev/nvidia0 195 0 uid=root"},
		{"uid overflows uint32", ".device c /dev/nvidia0 195 0 uid=99999999999"},
		{"gid not an integer", ".device c /dev/nvidia0 195 0 gid=wheel"},
		{"gid overflows uint32", ".device c /dev/nvidia0 195 0 gid=99999999999"},
		{"unknown keyword argument", ".device c /dev/nvidia0 195 0 owner=root"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sub, _ := newSub()
			err := sub.Build(parseGrant(t, tc.grant))
			if !errors.Is(err, ErrInvalidGrant) {
				t.Fatalf("err = %v, want ErrInvalidGrant", err)
			}
		})
	}
}

func TestKeyEmptyWhenPathMissing(t *testing.T) {
	sub, _ := newSub()
	g := &agl.Model{Args: []agl.Arg{{Type: agl.ArgName, Value: "c"}}}
	if got := sub.Key(g); got != "" {
		t.Fatalf("Key() = %q, want empty string", got)
	}
}
