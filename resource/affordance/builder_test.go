package affordance

import (
	"context"
	"errors"
	"testing"

	"github.com/cruciblehq/crux/affordance/rlimit"
	"github.com/cruciblehq/crux/manifest"
	"github.com/cruciblehq/crux/registry"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Fatal("NewBuilder returned nil")
	}

	// Every getter returns a live, non-nil accumulator.
	if b.Spec() == nil {
		t.Fatal("Spec() is nil")
	}
	if b.Provision() == nil {
		t.Fatal("Provision() is nil")
	}
	if b.Network() == nil {
		t.Fatal("Network() is nil")
	}
	if b.Volumes() == nil {
		t.Fatal("Volumes() is nil")
	}
	if b.Kernel() == nil {
		t.Fatal("Kernel() is nil")
	}

	// The section getters return the same instances wired into the spec.
	if b.Network() != b.Spec().Net {
		t.Error("Network() is not the spec's network section")
	}
	if b.Volumes() != b.Spec().Volume {
		t.Error("Volumes() is not the spec's volume section")
	}
	if b.Kernel() != b.Spec().Kernel {
		t.Error("Kernel() is not the spec's kernel section")
	}
}

func TestNewBuilderBaseline(t *testing.T) {
	b := NewBuilder()

	// A fresh builder has processed no grants, so every accumulator is empty.
	if k := b.Kernel(); len(k.Features) != 0 || len(k.Modules) != 0 || len(k.Versions) != 0 {
		t.Errorf("kernel baseline not empty: %+v", k)
	}
	if n := b.Network(); len(n.Ingress) != 0 || len(n.Egress) != 0 {
		t.Errorf("network baseline not empty: %+v", n)
	}
	if p := b.Provision(); p.CPU != 0 || p.Memory != 0 || p.Disk != 0 {
		t.Errorf("provision baseline not zero: %+v", p)
	}
}

func TestSubstituteGrant(t *testing.T) {
	// Nil params returns the grant unchanged.
	g := manifest.Grant{Source: "$name"}
	if got := substituteGrant(g, nil); got.Source != "$name" {
		t.Fatalf("nil params Source = %q, want unchanged", got.Source)
	}

	// Known params are replaced; unknown references are left intact.
	params := map[string]string{"port": "8080", "host": "db"}
	g = manifest.Grant{Source: "tcp://$host:$port/$missing"}
	got := substituteGrant(g, params)
	if got.Source != "tcp://db:8080/$missing" {
		t.Fatalf("Source = %q, want tcp://db:8080/$missing", got.Source)
	}
}

func TestBuildAllowsRepeatedMACHooks(t *testing.T) {
	b := NewBuilder()
	ctx := context.Background()
	grants := []manifest.Grant{
		{Source: `.mac file_open where file.path like "/var/lib/agent/**"`},
		{Source: `.mac file_open where task.uid = 0`},
	}
	src := registry.Source{}
	for _, grant := range grants {
		if err := b.Build(ctx, grant, src); err != nil {
			t.Fatalf("Build(%q): %v", grant.Source, err)
		}
	}
	if got := len(b.Spec().MAC.Rules); got != 2 {
		t.Fatalf("MAC rule count = %d, want 2", got)
	}
}

func TestBuildDeduplicatesRepeatedSeccompGrant(t *testing.T) {
	b := NewBuilder()
	ctx := context.Background()
	grant := manifest.Grant{Source: `.seccomp ioctl tty`}
	src := registry.Source{}
	if err := b.Build(ctx, grant, src); err != nil {
		t.Fatalf("first Build: %v", err)
	}
	count := len(b.Spec().OCI.Linux.Seccomp.Syscalls)
	if err := b.Build(ctx, grant, src); err != nil {
		t.Fatalf("second Build: %v", err)
	}
	if got := len(b.Spec().OCI.Linux.Seccomp.Syscalls); got != count {
		t.Fatalf("seccomp syscall count = %d, want %d", got, count)
	}
}

func TestBuildRepeatedRlimitSameValueNoOp(t *testing.T) {
	b := NewBuilder()
	ctx := context.Background()
	grant := manifest.Grant{Source: `.rlimit nofile 1024 1024`}
	src := registry.Source{}
	if err := b.Build(ctx, grant, src); err != nil {
		t.Fatalf("first Build: %v", err)
	}
	if err := b.Build(ctx, grant, src); err != nil {
		t.Fatalf("second Build: %v", err)
	}
}

func TestBuildRepeatedRlimitDifferentValueConflicts(t *testing.T) {
	b := NewBuilder()
	ctx := context.Background()
	src := registry.Source{}
	if err := b.Build(ctx, manifest.Grant{Source: `.rlimit nofile 1024 1024`}, src); err != nil {
		t.Fatalf("first Build: %v", err)
	}
	err := b.Build(ctx, manifest.Grant{Source: `.rlimit nofile 2048 2048`}, src)
	if !errors.Is(err, rlimit.ErrInvalidGrant) {
		t.Fatalf("err = %v, want rlimit.ErrInvalidGrant", err)
	}
}
