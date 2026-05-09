package registry

import "github.com/cruciblehq/crux/crex"

// Mutable properties of a version for creation or update.
//
// Used as the request body for version creation and update operations. For
// update requests, the version field must match the URL path parameter or
// update context. Contains only user-modifiable fields. The media type is
// [MediaTypeVersionInfo].
type VersionInfo struct {
	String string `json:"string"` // Version string (e.g., "1.0.0").
}

// Validates the version info.
//
// The version string must be a valid semantic version (see [ValidateVersionString]).
func (info *VersionInfo) Validate() error {
	if err := ValidateVersionString(info.String); err != nil {
		return crex.Wrap(ErrInvalidVersion, err)
	}
	return nil
}

// Lightweight version representation for listings.
//
// Provides version metadata without full archive details. Used in resource
// listings and version lists to keep payloads compact. Includes read-only
// fields like publication status and timestamps.
type VersionSummary struct {
	String    string `json:"string"`    // Version string (e.g., "1.0.0").
	CreatedAt int64  `json:"createdAt"` // When the version was created.
	UpdatedAt int64  `json:"updatedAt"` // When the version was last updated.
}

// Validates the version summary.
func (s *VersionSummary) Validate() error {
	if err := ValidateVersionString(s.String); err != nil {
		return crex.Wrap(ErrInvalidVersion, err)
	}
	if err := ValidateTimestamps(s.CreatedAt, s.UpdatedAt); err != nil {
		return crex.Wrap(ErrInvalidVersion, err)
	}
	return nil
}

// Complete version with archive details and publication status.
//
// Tracks both metadata (always mutable) and archive state (immutable after
// publication). The archive, size, and digest fields are null before archive
// upload and populated afterward. The publishedAt field is null for unpublished
// versions and contains the publication timestamp when published. Unpublished
// versions support archive replacement for iterative development, while
// published versions ensure immutability for stable dependency resolution.
// Version metadata updates remain allowed even after publication. Includes
// scoping information to identify the version's location. The media type is
// [MediaTypeVersion].
type Version struct {
	Namespace string  `json:"namespace"` // Namespace this version belongs to.
	Resource  string  `json:"resource"`  // Resource this version belongs to.
	String    string  `json:"string"`    // Version string (e.g., "1.0.0").
	Archive   *string `json:"archive"`   // Download URL or null if not uploaded.
	Size      *int64  `json:"size"`      // Archive size in bytes (null if not uploaded).
	Digest    *string `json:"digest"`    // Archive digest (e.g., "sha256:abc...", null if not uploaded).
	CreatedAt int64   `json:"createdAt"` // When the version was created.
	UpdatedAt int64   `json:"updatedAt"` // When the version was last updated.
}

// Validates the version.
func (v *Version) Validate() error {
	if err := ValidateName(v.Namespace); err != nil {
		return crex.Wrap(ErrInvalidVersion, err)
	}
	if err := ValidateName(v.Resource); err != nil {
		return crex.Wrap(ErrInvalidVersion, err)
	}
	if err := ValidateVersionString(v.String); err != nil {
		return crex.Wrap(ErrInvalidVersion, err)
	}
	if err := ValidateArchiveFields(v.Archive, v.Size, v.Digest); err != nil {
		return crex.Wrap(ErrInvalidVersion, err)
	}
	if err := ValidateTimestamps(v.CreatedAt, v.UpdatedAt); err != nil {
		return crex.Wrap(ErrInvalidVersion, err)
	}
	return nil
}

// Collection of versions for a resource.
//
// The media type is [MediaTypeVersionList].
type VersionList struct {

	// List of versions.
	Versions []VersionSummary `json:"versions"`
}

// Validates the version list.
func (l *VersionList) Validate() error {
	for i := range l.Versions {
		if err := l.Versions[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidVersion, err)
		}
	}
	return nil
}
