package registry

import (
	"errors"
	"testing"
)

func TestNamespaceInfo_Validate_Valid(t *testing.T) {
	info := NamespaceInfo{Name: "my-namespace"}
	if err := info.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNamespaceInfo_Validate_Invalid_Name(t *testing.T) {
	info := NamespaceInfo{Name: ""}
	err := info.Validate()
	if !errors.Is(err, ErrInvalidNamespace) {
		t.Errorf("expected ErrInvalidNamespace, got %v", err)
	}
}

func TestNamespaceSummary_Validate_Valid(t *testing.T) {
	s := NamespaceSummary{
		Name:          "my-namespace",
		ResourceCount: 0,
		CreatedAt:     1,
		UpdatedAt:     1,
	}
	if err := s.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNamespaceSummary_Validate_Invalid_Name(t *testing.T) {
	s := NamespaceSummary{
		Name:      "",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidNamespace) {
		t.Errorf("expected ErrInvalidNamespace, got %v", err)
	}
}

func TestNamespaceSummary_Validate_Invalid_NegativeCount(t *testing.T) {
	s := NamespaceSummary{
		Name:          "my-namespace",
		ResourceCount: -1,
		CreatedAt:     1,
		UpdatedAt:     1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidNamespace) {
		t.Errorf("expected ErrInvalidNamespace, got %v", err)
	}
}

func TestNamespaceSummary_Validate_Invalid_Timestamps(t *testing.T) {
	s := NamespaceSummary{
		Name:      "my-namespace",
		CreatedAt: 0,
		UpdatedAt: 1,
	}
	err := s.Validate()
	if !errors.Is(err, ErrInvalidNamespace) {
		t.Errorf("expected ErrInvalidNamespace, got %v", err)
	}
}

func TestNamespace_Validate_Valid(t *testing.T) {
	ns := Namespace{
		Name:      "my-namespace",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	if err := ns.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNamespace_Validate_Valid_WithResources(t *testing.T) {
	ns := Namespace{
		Name: "my-namespace",
		Resources: []ResourceSummary{
			{Name: "my-resource", Type: "widget", CreatedAt: 1, UpdatedAt: 1},
		},
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	if err := ns.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNamespace_Validate_Invalid_Name(t *testing.T) {
	ns := Namespace{
		Name:      "",
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := ns.Validate()
	if !errors.Is(err, ErrInvalidNamespace) {
		t.Errorf("expected ErrInvalidNamespace, got %v", err)
	}
}

func TestNamespace_Validate_Invalid_Timestamps(t *testing.T) {
	ns := Namespace{
		Name:      "my-namespace",
		CreatedAt: 2,
		UpdatedAt: 1,
	}
	err := ns.Validate()
	if !errors.Is(err, ErrInvalidNamespace) {
		t.Errorf("expected ErrInvalidNamespace, got %v", err)
	}
}

func TestNamespace_Validate_Invalid_NestedResource(t *testing.T) {
	ns := Namespace{
		Name: "my-namespace",
		Resources: []ResourceSummary{
			{Name: "", Type: "widget", CreatedAt: 1, UpdatedAt: 1},
		},
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	err := ns.Validate()
	if !errors.Is(err, ErrInvalidNamespace) {
		t.Errorf("expected ErrInvalidNamespace, got %v", err)
	}
}

func TestNamespaceList_Validate_Valid_Empty(t *testing.T) {
	l := NamespaceList{}
	if err := l.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNamespaceList_Validate_Valid_NonEmpty(t *testing.T) {
	l := NamespaceList{
		Namespaces: []NamespaceSummary{
			{Name: "my-namespace", CreatedAt: 1, UpdatedAt: 1},
		},
	}
	if err := l.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNamespaceList_Validate_Invalid_NestedNamespace(t *testing.T) {
	l := NamespaceList{
		Namespaces: []NamespaceSummary{
			{Name: "", CreatedAt: 1, UpdatedAt: 1},
		},
	}
	err := l.Validate()
	if !errors.Is(err, ErrInvalidNamespace) {
		t.Errorf("expected ErrInvalidNamespace, got %v", err)
	}
}
