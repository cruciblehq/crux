package registry

import (
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "not found error",
			err:  &Error{Code: ErrorCodeNotFound, Message: "resource not found"},
			want: "not_found: resource not found",
		},
		{
			name: "bad request error",
			err:  &Error{Code: ErrorCodeBadRequest, Message: "invalid input"},
			want: "bad_request: invalid input",
		},
		{
			name: "internal error",
			err:  &Error{Code: ErrorCodeInternalError, Message: "database connection failed"},
			want: "internal_error: database connection failed",
		},
		{
			name: "namespace exists error",
			err:  &Error{Code: ErrorCodeNamespaceExists, Message: "namespace already exists"},
			want: "namespace_exists: namespace already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMediaType_Constants(t *testing.T) {
	// Verify all media types follow the correct pattern
	mediaTypes := map[string]MediaType{
		"error":          MediaTypeError,
		"namespace-info": MediaTypeNamespaceInfo,
		"namespace":      MediaTypeNamespace,
		"namespace-list": MediaTypeNamespaceList,
		"resource-info":  MediaTypeResourceInfo,
		"resource":       MediaTypeResource,
		"resource-list":  MediaTypeResourceList,
		"version-info":   MediaTypeVersionInfo,
		"version":        MediaTypeVersion,
		"version-list":   MediaTypeVersionList,
		"channel-info":   MediaTypeChannelInfo,
		"channel":        MediaTypeChannel,
		"channel-list":   MediaTypeChannelList,
		"archive":        MediaTypeArchive,
	}

	expectedPrefix := "application/vnd.crucible."
	expectedSuffix := ".v0"

	for name, mt := range mediaTypes {
		t.Run(name, func(t *testing.T) {
			s := string(mt)

			// Check prefix
			if len(s) < len(expectedPrefix) || s[:len(expectedPrefix)] != expectedPrefix {
				t.Errorf("MediaType %q does not start with %q", s, expectedPrefix)
			}

			// Check suffix
			if len(s) < len(expectedSuffix) || s[len(s)-len(expectedSuffix):] != expectedSuffix {
				t.Errorf("MediaType %q does not end with %q", s, expectedSuffix)
			}

			// Check it contains the name
			if len(s) > len(expectedPrefix)+len(expectedSuffix) {
				middle := s[len(expectedPrefix) : len(s)-len(expectedSuffix)]
				if middle != name {
					t.Errorf("MediaType %q middle section = %q, want %q", s, middle, name)
				}
			}
		})
	}
}

func TestErrorCode_Constants(t *testing.T) {
	// Verify all error codes are defined and follow snake_case convention
	errorCodes := []ErrorCode{
		ErrorCodeBadRequest,
		ErrorCodeNotFound,
		ErrorCodeNamespaceExists,
		ErrorCodeNamespaceNotEmpty,
		ErrorCodeResourceExists,
		ErrorCodeResourceHasPublished,
		ErrorCodeVersionExists,
		ErrorCodeVersionPublished,
		ErrorCodeChannelExists,
		ErrorCodePreconditionFailed,
		ErrorCodeInternalError,
	}

	for _, code := range errorCodes {
		t.Run(string(code), func(t *testing.T) {
			s := string(code)
			if s == "" {
				t.Error("error code is empty")
			}

			// Verify it's lowercase with underscores
			for _, ch := range s {
				if !((ch >= 'a' && ch <= 'z') || ch == '_') {
					t.Errorf("error code %q contains invalid character %q", s, ch)
				}
			}
		})
	}
}

func TestNamespaceInfo_Fields(t *testing.T) {
	info := NamespaceInfo{
		Name:        "test-namespace",
		Description: "Test description",
	}

	if info.Name != "test-namespace" {
		t.Errorf("Name = %q, want 'test-namespace'", info.Name)
	}
	if info.Description != "Test description" {
		t.Errorf("Description = %q, want 'Test description'", info.Description)
	}
}

func TestResourceInfo_Fields(t *testing.T) {
	info := ResourceInfo{
		Name:        "test-resource",
		Type:        "widget",
		Description: "Test resource description",
	}

	if info.Name != "test-resource" {
		t.Errorf("Name = %q, want 'test-resource'", info.Name)
	}
	if info.Type != "widget" {
		t.Errorf("Type = %q, want 'widget'", info.Type)
	}
	if info.Description != "Test resource description" {
		t.Errorf("Description = %q, want 'Test resource description'", info.Description)
	}
}

func TestVersionInfo_Fields(t *testing.T) {
	info := VersionInfo{
		String: "1.2.3",
	}

	if info.String != "1.2.3" {
		t.Errorf("String = %q, want '1.2.3'", info.String)
	}
}

func TestChannelInfo_Fields(t *testing.T) {
	info := ChannelInfo{
		Name:        "stable",
		Version:     "1.0.0",
		Description: "Stable channel",
	}

	if info.Name != "stable" {
		t.Errorf("Name = %q, want 'stable'", info.Name)
	}
	if info.Version != "1.0.0" {
		t.Errorf("Version = %q, want '1.0.0'", info.Version)
	}
	if info.Description != "Stable channel" {
		t.Errorf("Description = %q, want 'Stable channel'", info.Description)
	}
}

func TestVersion_NullableFields(t *testing.T) {
	// Test version without archive
	v1 := Version{
		Namespace: "test-ns",
		Resource:  "test-resource",
		String:    "1.0.0",
		Digest:    nil,
		Size:      nil,
		Archive:   nil,
	}

	if v1.Digest != nil {
		t.Error("Digest should be nil for version without archive")
	}
	if v1.Size != nil {
		t.Error("Size should be nil for version without archive")
	}
	if v1.Archive != nil {
		t.Error("Archive should be nil for version without archive")
	}

	// Test version with archive
	digest := "abc123"
	size := int64(1024)
	path := "/path/to/archive.tar.zst"

	v2 := Version{
		Namespace: "test-ns",
		Resource:  "test-resource",
		String:    "2.0.0",
		Digest:    &digest,
		Size:      &size,
		Archive:   &path,
	}

	if v2.Digest == nil || *v2.Digest != "abc123" {
		t.Errorf("Digest = %v, want 'abc123'", v2.Digest)
	}
	if v2.Size == nil || *v2.Size != 1024 {
		t.Errorf("Size = %v, want 1024", v2.Size)
	}
	if v2.Archive == nil || *v2.Archive != "/path/to/archive.tar.zst" {
		t.Errorf("Archive = %v, want '/path/to/archive.tar.zst'", v2.Archive)
	}
}

func TestNamespaceSummary_Fields(t *testing.T) {
	summary := NamespaceSummary{
		Name:          "test-ns",
		Description:   "Test namespace",
		ResourceCount: 5,
		CreatedAt:     1234567890,
		UpdatedAt:     1234567900,
	}

	if summary.Name != "test-ns" {
		t.Errorf("Name = %q, want 'test-ns'", summary.Name)
	}
	if summary.ResourceCount != 5 {
		t.Errorf("ResourceCount = %d, want 5", summary.ResourceCount)
	}
	if summary.CreatedAt != 1234567890 {
		t.Errorf("CreatedAt = %d, want 1234567890", summary.CreatedAt)
	}
	if summary.UpdatedAt != 1234567900 {
		t.Errorf("UpdatedAt = %d, want 1234567900", summary.UpdatedAt)
	}
}

func TestResourceSummary_Fields(t *testing.T) {
	latestVersion := "1.2.3"
	summary := ResourceSummary{
		Name:          "test-resource",
		Type:          "widget",
		Description:   "Test resource",
		LatestVersion: &latestVersion,
		VersionCount:  10,
		ChannelCount:  3,
		CreatedAt:     1234567890,
		UpdatedAt:     1234567900,
	}

	if summary.Name != "test-resource" {
		t.Errorf("Name = %q, want 'test-resource'", summary.Name)
	}
	if summary.Type != "widget" {
		t.Errorf("Type = %q, want 'widget'", summary.Type)
	}
	if summary.VersionCount != 10 {
		t.Errorf("VersionCount = %d, want 10", summary.VersionCount)
	}
	if summary.ChannelCount != 3 {
		t.Errorf("ChannelCount = %d, want 3", summary.ChannelCount)
	}
	if summary.LatestVersion == nil || *summary.LatestVersion != "1.2.3" {
		t.Errorf("LatestVersion = %v, want '1.2.3'", summary.LatestVersion)
	}
}

func TestVersionSummary_Fields(t *testing.T) {
	summary := VersionSummary{
		String:    "1.2.3",
		CreatedAt: 1234567890,
		UpdatedAt: 1234567900,
	}

	if summary.String != "1.2.3" {
		t.Errorf("String = %q, want '1.2.3'", summary.String)
	}
	if summary.CreatedAt != 1234567890 {
		t.Errorf("CreatedAt = %d, want 1234567890", summary.CreatedAt)
	}
	if summary.UpdatedAt != 1234567900 {
		t.Errorf("UpdatedAt = %d, want 1234567900", summary.UpdatedAt)
	}
}

func TestChannelSummary_Fields(t *testing.T) {
	summary := ChannelSummary{
		Name:        "stable",
		Version:     "1.0.0",
		Description: "Stable channel",
		CreatedAt:   1234567890,
		UpdatedAt:   1234567900,
	}

	if summary.Name != "stable" {
		t.Errorf("Name = %q, want 'stable'", summary.Name)
	}
	if summary.Version != "1.0.0" {
		t.Errorf("Version = %q, want '1.0.0'", summary.Version)
	}
	if summary.Description != "Stable channel" {
		t.Errorf("Description = %q, want 'Stable channel'", summary.Description)
	}
	if summary.CreatedAt != 1234567890 {
		t.Errorf("CreatedAt = %d, want 1234567890", summary.CreatedAt)
	}
}
