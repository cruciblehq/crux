package fcap

import (
	"errors"
	"testing"
)

func TestParseSingleCapabilityGrant(t *testing.T) {
	g, err := Parse("/usr/bin/ping effective net_raw")
	if err != nil {
		t.Fatal(err)
	}
	if g.Mode != ModeEffective {
		t.Fatalf("mode = %q, want %q", g.Mode, ModeEffective)
	}
	if g.Path != "/usr/bin/ping" {
		t.Fatalf("path = %q, want %q", g.Path, "/usr/bin/ping")
	}
	if len(g.Caps) != 1 || g.Caps[0] != "net_raw" {
		t.Fatalf("caps = %#v", g.Caps)
	}
}

func TestParseAcceptsInheritableMode(t *testing.T) {
	g, err := Parse("/usr/bin/ping inheritable net_raw")
	if err != nil {
		t.Fatal(err)
	}
	if g.Mode != ModeInheritable {
		t.Fatalf("mode = %q, want %q", g.Mode, ModeInheritable)
	}
}

func TestParseRejectsMissingFields(t *testing.T) {
	_, err := Parse("/usr/bin/ping effective")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRejectsUnknownMode(t *testing.T) {
	_, err := Parse("/usr/bin/ping bogus net_raw")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRejectsRelativePath(t *testing.T) {
	_, err := Parse("usr/bin/ping effective net_raw")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRejectsNonCleanEquivalentPath(t *testing.T) {
	_, err := Parse("/usr//bin///ping effective net_raw")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRejectsDotSegmentsInPath(t *testing.T) {
	_, err := Parse("/usr/./bin/ping effective net_raw")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRejectsDotDotSegmentsInPath(t *testing.T) {
	_, err := Parse("/usr/bin/../sbin/ping effective net_raw")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRejectsDoubleSlashPrefix(t *testing.T) {
	_, err := Parse("//usr/bin/ping effective net_raw")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRejectsTrailingSlashPath(t *testing.T) {
	_, err := Parse("/usr/bin/ping/ effective net_raw")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRejectsUnknownCapability(t *testing.T) {
	_, err := Parse("/usr/bin/ping effective not_a_cap")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestParseRejectsMultipleCapabilities(t *testing.T) {
	_, err := Parse("/usr/bin/ping effective net_raw net_admin")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestNormalizePathAcceptsAbsolutePath(t *testing.T) {
	got, err := normalizePath("/usr/bin/ping")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/usr/bin/ping" {
		t.Fatalf("path = %q, want %q", got, "/usr/bin/ping")
	}
}

func TestNormalizePathRejectsNonCleanEquivalentPath(t *testing.T) {
	_, err := normalizePath("/usr//bin///ping")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestNormalizePathRejectsDotSegment(t *testing.T) {
	_, err := normalizePath("/usr/./bin/ping")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestNormalizePathRejectsDotDotSegment(t *testing.T) {
	_, err := normalizePath("/usr/bin/../sbin/ping")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestNormalizePathRejectsEmptyPath(t *testing.T) {
	_, err := normalizePath("   ")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestNormalizePathRejectsRelativePath(t *testing.T) {
	_, err := normalizePath("usr/bin/ping")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestNormalizePathRejectsTrailingSlash(t *testing.T) {
	_, err := normalizePath("/usr/bin/ping/")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestNormalizePathRejectsDoubleSlashPrefix(t *testing.T) {
	_, err := normalizePath("//usr/bin/ping")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestNormalizePathRejectsNUL(t *testing.T) {
	_, err := normalizePath("/usr/bin/ping\x00")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}

func TestNormalizePathRejectsRoot(t *testing.T) {
	_, err := normalizePath("/")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}
