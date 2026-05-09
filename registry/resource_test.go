package registry

import (
	"errors"
	"testing"
)

func TestResourceInfo_Validate_Valid(t *testing.T) {
	info := ResourceInfo{Name: "my-resource", Type: "widget"}
	if err := info.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResourceInfo_Validate_Invalid_Name(t *testing.T) {
	info := ResourceInfo{Name: "", Type: "widget"}
	err := info.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResourceInfo_Validate_Invalid_Type(t *testing.T) {
	info := ResourceInfo{Name: "my-resource", Type: ""}
	err := info.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResourceSummary_Validate_Valid(t *testing.T) {
	s := ResourceSummary{
		Name:      "my-resource",
		Type:      "widget",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	if err := s.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResourceSummary_Validate_Valid_WithLatestVersion(t *testing.T) {
	latest := "1.0.0"
	s := ResourceSummary{
		Name:          "my-resource",
		Type:          "widget",
		LatestVersion: &latest,
		CreatedAt:     1,
		UpdatedAt:     1,
	}
	if err := s.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResourceSummary_Validate_Invalid_Name(t *testing.T) {
	s := ResourceSummary{
		Name:      "",
		Type:      "widget",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResourceSummary_Validate_Invalid_Type(t *testing.T) {
	s := ResourceSummary{
		Name:      "my-resource",
		Type:      "",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResourceSummary_Validate_Invalid_LatestVersion(t *testing.T) {
	latest := "not-a-version"
	s := ResourceSummary{
		Name:          "my-resource",
		Type:          "widget",
		LatestVersion: &latest,
		CreatedAt:     1,
		UpdatedAt:     1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResourceSummary_Validate_Invalid_NegativeVersionCount(t *testing.T) {
	s := ResourceSummary{
		Name:         "my-resource",
		Type:         "widget",
		VersionCount: -1,
		CreatedAt:    1,
		UpdatedAt:    1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResourceSummary_Validate_Invalid_NegativeChannelCount(t *testing.T) {
	s := ResourceSummary{
		Name:         "my-resource",
		Type:         "widget",
		ChannelCount: -1,
		CreatedAt:    1,
		UpdatedAt:    1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResourceSummary_Validate_Invalid_Timestamps(t *testing.T) {
	s := ResourceSummary{
		Name:      "my-resource",
		Type:      "widget",
		CreatedAt: 0,
		UpdatedAt: 1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResource_Validate_Valid(t *testing.T) {
	r := Resource{
		Namespace: "my-namespace",
		Name:      "my-resource",
		Type:      "widget",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	if err := r.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResource_Validate_Valid_WithVersionsAndChannels(t *testing.T) {
	r := Resource{
		Namespace: "my-namespace",
		Name:      "my-resource",
		Type:      "widget",
		Versions: []VersionSummary{
			{String: "1.0.0", CreatedAt: 1, UpdatedAt: 1},
		},
		Channels: []ChannelSummary{
			{Name: "stable", Version: "1.0.0", CreatedAt: 1, UpdatedAt: 1},
		},
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	if err := r.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResource_Validate_Invalid_Namespace(t *testing.T) {
	r := Resource{
		Namespace: "",
		Name:      "my-resource",
		Type:      "widget",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := r.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResource_Validate_Invalid_Name(t *testing.T) {
	r := Resource{
		Namespace: "my-namespace",
		Name:      "",
		Type:      "widget",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := r.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResource_Validate_Invalid_Type(t *testing.T) {
	r := Resource{
		Namespace: "my-namespace",
		Name:      "my-resource",
		Type:      "",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := r.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResource_Validate_Invalid_Timestamps(t *testing.T) {
	r := Resource{
		Namespace: "my-namespace",
		Name:      "my-resource",
		Type:      "widget",
		CreatedAt: 2,
		UpdatedAt: 1,
	}
	err := r.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResource_Validate_Invalid_NestedVersion(t *testing.T) {
	r := Resource{
		Namespace: "my-namespace",
		Name:      "my-resource",
		Type:      "widget",
		Versions: []VersionSummary{
			{String: "bad", CreatedAt: 1, UpdatedAt: 1},
		},
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := r.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResource_Validate_Invalid_NestedChannel(t *testing.T) {
	r := Resource{
		Namespace: "my-namespace",
		Name:      "my-resource",
		Type:      "widget",
		Channels: []ChannelSummary{
			{Name: "", Version: "1.0.0", CreatedAt: 1, UpdatedAt: 1},
		},
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := r.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}

func TestResourceList_Validate_Valid_Empty(t *testing.T) {
	l := ResourceList{}
	if err := l.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResourceList_Validate_Valid_NonEmpty(t *testing.T) {
	l := ResourceList{
		Resources: []ResourceSummary{
			{Name: "my-resource", Type: "widget", CreatedAt: 1, UpdatedAt: 1},
		},
	}
	if err := l.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResourceList_Validate_Invalid_NestedResource(t *testing.T) {
	l := ResourceList{
		Resources: []ResourceSummary{
			{Name: "", Type: "widget", CreatedAt: 1, UpdatedAt: 1},
		},
	}
	err := l.Validate()
	if !errors.Is(err, ErrInvalidResource) {
		t.Errorf("expected ErrInvalidResource, got %v", err)
	}
}
