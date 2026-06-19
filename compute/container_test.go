package compute

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/containerd/containerd/v2/core/mount"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func TestParseID(t *testing.T) {
	tests := []struct {
		input   string
		want    uint32
		wantErr bool
	}{
		{"0", 0, false},
		{"1000", 1000, false},
		{"2147483647", 2147483647, false},
		{"2147483648", 0, true}, // exceeds runc maximum
		{"4294967295", 0, true}, // exceeds runc maximum
		{"-1", 0, true},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tc := range tests {
		got, err := parseID(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("parseID(%q) error = %v; wantErr %v", tc.input, err, tc.wantErr)
			continue
		}
		if !tc.wantErr && got != tc.want {
			t.Errorf("parseID(%q) = %d; want %d", tc.input, got, tc.want)
		}
	}
}

func TestParseUserSpec(t *testing.T) {
	u32 := func(v uint32) *uint32 { return &v }

	tests := []struct {
		input   string
		wantUID uint32
		wantGID *uint32
		wantErr bool
	}{
		{"0", 0, nil, false},
		{"1000", 1000, nil, false},
		{"1000:2000", 1000, u32(2000), false},
		{"0:0", 0, u32(0), false},
		{"2147483647:2147483647", 2147483647, u32(2147483647), false},
		{"abc", 0, nil, true},
		{"1000:abc", 0, nil, true},
		{"2147483648", 0, nil, true},
		{"1000:2147483648", 0, nil, true},
	}

	for _, tc := range tests {
		uid, gid, err := parseUserSpec(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("parseUserSpec(%q) error = %v; wantErr %v", tc.input, err, tc.wantErr)
			continue
		}
		if tc.wantErr {
			continue
		}
		if uid != tc.wantUID {
			t.Errorf("parseUserSpec(%q) uid = %d; want %d", tc.input, uid, tc.wantUID)
		}
		if tc.wantGID == nil && gid != nil {
			t.Errorf("parseUserSpec(%q) gid = %v; want nil", tc.input, gid)
		} else if tc.wantGID != nil && (gid == nil || *gid != *tc.wantGID) {
			t.Errorf("parseUserSpec(%q) gid = %v; want %d", tc.input, gid, *tc.wantGID)
		}
	}
}

func baseSpec(shell string, env []string, cwd string) *specs.Spec {
	return &specs.Spec{
		Process: &specs.Process{
			Args:     []string{shell, "-c", "true"},
			Env:      env,
			Cwd:      cwd,
			Terminal: true,
			User:     specs.User{UID: 0, GID: 0},
		},
	}
}

func TestBuildExecProcess_DefaultShell(t *testing.T) {
	spec := baseSpec("/bin/sh", nil, "/")
	proc := buildExecProcess(spec, "echo hello", &RuntimeOptions{})
	if len(proc.Args) != 3 || proc.Args[0] != "/bin/sh" || proc.Args[1] != "-c" || proc.Args[2] != "echo hello" {
		t.Errorf("unexpected args: %v", proc.Args)
	}
	if proc.Terminal {
		t.Error("Terminal should be false")
	}
}

func TestBuildExecProcess_CustomShell(t *testing.T) {
	spec := baseSpec("/bin/sh", nil, "/")
	proc := buildExecProcess(spec, "echo hello", &RuntimeOptions{Shell: "/bin/bash"})
	if proc.Args[0] != "/bin/bash" {
		t.Errorf("expected /bin/bash, got %s", proc.Args[0])
	}
}

func TestBuildExecProcess_EnvMerge(t *testing.T) {
	spec := baseSpec("/bin/sh", []string{"A=1", "B=2"}, "/")
	proc := buildExecProcess(spec, "x", &RuntimeOptions{Env: map[string]string{"A": "99", "C": "3"}})
	found := map[string]string{}
	for _, e := range proc.Env {
		if k, v, ok := strings.Cut(e, "="); ok {
			found[k] = v
		}
	}
	if found["A"] != "99" {
		t.Errorf("A should be overridden to 99, got %q", found["A"])
	}
	if found["B"] != "2" {
		t.Errorf("B should remain 2, got %q", found["B"])
	}
	if found["C"] != "3" {
		t.Errorf("C should be added as 3, got %q", found["C"])
	}
}

func TestBuildExecProcess_Workdir(t *testing.T) {
	spec := baseSpec("/bin/sh", nil, "/")
	proc := buildExecProcess(spec, "x", &RuntimeOptions{Workdir: "/tmp"})
	if proc.Cwd != "/tmp" {
		t.Errorf("expected Cwd /tmp, got %s", proc.Cwd)
	}
}

func TestBuildExecProcess_WorkdirEmpty(t *testing.T) {
	spec := baseSpec("/bin/sh", nil, "/home")
	proc := buildExecProcess(spec, "x", &RuntimeOptions{})
	if proc.Cwd != "/home" {
		t.Errorf("expected Cwd to be inherited as /home, got %s", proc.Cwd)
	}
}

func TestBuildExecProcess_User(t *testing.T) {
	spec := baseSpec("/bin/sh", nil, "/")
	proc := buildExecProcess(spec, "x", &RuntimeOptions{User: "1000:2000"})
	if proc.User.UID != 1000 {
		t.Errorf("expected UID 1000, got %d", proc.User.UID)
	}
	if proc.User.GID != 2000 {
		t.Errorf("expected GID 2000, got %d", proc.User.GID)
	}
}

func TestBuildExecProcess_InvalidUser(t *testing.T) {
	spec := baseSpec("/bin/sh", nil, "/")
	spec.Process.User = specs.User{UID: 5, GID: 6}
	proc := buildExecProcess(spec, "x", &RuntimeOptions{User: "notanumber"})
	// Invalid user spec is silently ignored; original UID/GID unchanged.
	if proc.User.UID != 5 || proc.User.GID != 6 {
		t.Errorf("expected original UID/GID to be preserved, got %d/%d", proc.User.UID, proc.User.GID)
	}
}

func TestBuildExecProcess_DoesNotMutateBase(t *testing.T) {
	spec := baseSpec("/bin/sh", []string{"A=1"}, "/")
	origArgs := append([]string{}, spec.Process.Args...)
	origEnv := append([]string{}, spec.Process.Env...)

	buildExecProcess(spec, "x", &RuntimeOptions{Env: map[string]string{"B": "2"}})

	if !reflect.DeepEqual(spec.Process.Args, origArgs) {
		t.Error("buildExecProcess mutated base spec args")
	}
	if !reflect.DeepEqual(spec.Process.Env, origEnv) {
		t.Error("buildExecProcess mutated base spec env")
	}
}

func TestOverlayHasChanges_EmptyUpperdir(t *testing.T) {
	dir := t.TempDir()
	mounts := []mount.Mount{
		{Type: overlayMountType, Options: []string{overlayUpperdirOpt + dir}},
	}
	if overlayHasChanges(mounts) {
		t.Error("expected no changes in empty upperdir")
	}
}

func TestOverlayHasChanges_NonEmptyUpperdir(t *testing.T) {
	dir := t.TempDir()
	f, err := os.Create(filepath.Join(dir, "file"))
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	mounts := []mount.Mount{
		{Type: overlayMountType, Options: []string{overlayUpperdirOpt + dir}},
	}
	if !overlayHasChanges(mounts) {
		t.Error("expected changes with a file in upperdir")
	}
}

func TestOverlayHasChanges_MissingUpperdir(t *testing.T) {
	mounts := []mount.Mount{
		{Type: overlayMountType, Options: []string{overlayUpperdirOpt + "/nonexistent/path/xyz"}},
	}
	// ReadDir error is treated as conservative true.
	if !overlayHasChanges(mounts) {
		t.Error("expected true for missing upperdir (conservative)")
	}
}

func TestOverlayHasChanges_NoOverlayMount(t *testing.T) {
	mounts := []mount.Mount{
		{Type: "tmpfs", Options: []string{}},
	}
	if overlayHasChanges(mounts) {
		t.Error("expected false when no overlay mount is present")
	}
}

func TestOverlayHasChanges_NoUpperdirOption(t *testing.T) {
	mounts := []mount.Mount{
		{Type: overlayMountType, Options: []string{"lowerdir=/lower"}},
	}
	if overlayHasChanges(mounts) {
		t.Error("expected false when upperdir option is absent")
	}
}
