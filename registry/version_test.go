package registry

import (
	"errors"
	"testing"
)

func TestVersionInfo_Validate_Valid(t *testing.T) {
	info := VersionInfo{String: "1.0.0"}
	if err := info.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVersionInfo_Validate_Invalid(t *testing.T) {
	info := VersionInfo{String: "not-a-version"}
	err := info.Validate()
	if !errors.Is(err, ErrInvalidVersion) {
		t.Errorf("expected ErrInvalidVersion, got %v", err)
	}
}

func TestVersionSummary_Validate_Valid(t *testing.T) {
	s := VersionSummary{
		String:    "1.0.0",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	if err := s.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVersionSummary_Validate_Invalid_String(t *testing.T) {
	s := VersionSummary{
		String:    "bad",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidVersion) {
		t.Errorf("expected ErrInvalidVersion, got %v", err)
	}
}

func TestVersionSummary_Validate_Invalid_Timestamps(t *testing.T) {
	s := VersionSummary{
		String:    "1.0.0",
		CreatedAt: 0,
		UpdatedAt: 1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidVersion) {
		t.Errorf("expected ErrInvalidVersion, got %v", err)
	}
}

func TestVersion_Validate_Valid_NoArchive(t *testing.T) {
	v := Version{
		Namespace: "my-namespace",
		Resource:  "my-resource",
		String:    "1.0.0",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	if err := v.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVersion_Validate_Valid_WithArchive(t *testing.T) {
	archive := "https://example.com/archive.tar.zst"
	size := int64(1024)
	digest := "sha256:abc123"
	v := Version{
		Namespace: "my-namespace",
		Resource:  "my-resource",
		String:    "1.0.0",
		Archive:   &archive,
		Size:      &size,
		Digest:    &digest,
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	if err := v.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVersion_Validate_Invalid_Namespace(t *testing.T) {
	v := Version{
		Namespace: "",
		Resource:  "my-resource",
		String:    "1.0.0",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := v.Validate()
	if !errors.Is(err, ErrInvalidVersion) {
		t.Errorf("expected ErrInvalidVersion, got %v", err)
	}
}

func TestVersion_Validate_Invalid_Resource(t *testing.T) {
	v := Version{
		Namespace: "my-namespace",
		Resource:  "",
		String:    "1.0.0",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := v.Validate()
	if !errors.Is(err, ErrInvalidVersion) {
		t.Errorf("expected ErrInvalidVersion, got %v", err)
	}
}

func TestVersion_Validate_Invalid_String(t *testing.T) {
	v := Version{
		Namespace: "my-namespace",
		Resource:  "my-resource",
		String:    "bad",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := v.Validate()
	if !errors.Is(err, ErrInvalidVersion) {
		t.Errorf("expected ErrInvalidVersion, got %v", err)
	}
}

func TestVersion_Validate_Invalid_IncompleteArchive(t *testing.T) {
	archive := "https://example.com/archive.tar.zst"
	v := Version{
		Namespace: "my-namespace",
		Resource:  "my-resource",
		String:    "1.0.0",
		Archive:   &archive,
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := v.Validate()
	if !errors.Is(err, ErrInvalidVersion) {
		t.Errorf("expected ErrInvalidVersion, got %v", err)
	}
}

func TestVersion_Validate_Invalid_Timestamps(t *testing.T) {
	v := Version{
		Namespace: "my-namespace",
		Resource:  "my-resource",
		String:    "1.0.0",
		CreatedAt: 2,
		UpdatedAt: 1,
	}
	err := v.Validate()
	if !errors.Is(err, ErrInvalidVersion) {
		t.Errorf("expected ErrInvalidVersion, got %v", err)
	}
}

func TestVersionList_Validate_Valid_Empty(t *testing.T) {
	l := VersionList{}
	if err := l.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVersionList_Validate_Valid_NonEmpty(t *testing.T) {
	l := VersionList{
		Versions: []VersionSummary{
			{String: "1.0.0", CreatedAt: 1, UpdatedAt: 1},
		},
	}
	if err := l.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVersionList_Validate_Invalid_NestedVersion(t *testing.T) {
	l := VersionList{
		Versions: []VersionSummary{
			{String: "bad", CreatedAt: 1, UpdatedAt: 1},
		},
	}
	err := l.Validate()
	if !errors.Is(err, ErrInvalidVersion) {
		t.Errorf("expected ErrInvalidVersion, got %v", err)
	}
}
