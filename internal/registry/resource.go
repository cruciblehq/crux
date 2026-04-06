package registry

import "github.com/cruciblehq/crex"

// Mutable properties of a resource for creation or update.
//
// Used as the request body for resource creation and update operations. For
// update requests, the name field must match the URL path parameter or update
// context. Contains only user-modifiable fields. The media type is
// [MediaTypeResourceInfo].
type ResourceInfo struct {
	Name        string `codec:"name"`        // Resource name.
	Type        string `codec:"type"`        // Resource type (e.g., "widget", "service").
	Description string `codec:"description"` // Description.
}

// Validates the resource info.
//
// The name must conform to the registry naming rules (see [ValidateName])
// and the type must not be empty.
func (info *ResourceInfo) Validate() error {
	if err := ValidateName(info.Name); err != nil {
		return crex.Wrap(ErrInvalidResource, err)
	}
	if err := ValidateResourceType(info.Type); err != nil {
		return crex.Wrap(ErrInvalidResource, err)
	}
	return nil
}

// Lightweight resource representation for listings.
//
// Provides resource metadata without nested version and channel details. Used
// in namespace listings and resource lists to keep payloads compact. Includes
// read-only fields like timestamps, statistics, and latest version information
// to help with navigation decisions.
type ResourceSummary struct {
	Name          string  `codec:"name"`          // Resource name.
	Type          string  `codec:"type"`          // Resource type (e.g., "widget", "service").
	Description   string  `codec:"description"`   // Description.
	LatestVersion *string `codec:"latestVersion"` // Most recent version string (null if no versions).
	VersionCount  int     `codec:"versionCount"`  // Number of versions for this resource.
	ChannelCount  int     `codec:"channelCount"`  // Number of channels for this resource.
	CreatedAt     int64   `codec:"createdAt"`     // When the resource was created.
	UpdatedAt     int64   `codec:"updatedAt"`     // When the resource was last updated.
}

// Validates the resource summary.
func (s *ResourceSummary) Validate() error {
	if err := ValidateName(s.Name); err != nil {
		return crex.Wrap(ErrInvalidResource, err)
	}
	if err := ValidateResourceType(s.Type); err != nil {
		return crex.Wrap(ErrInvalidResource, err)
	}
	if s.LatestVersion != nil {
		if err := ValidateVersionString(*s.LatestVersion); err != nil {
			return crex.Wrap(ErrInvalidResource, err)
		}
	}
	if err := ValidateCount(s.VersionCount); err != nil {
		return crex.Wrap(ErrInvalidResource, err)
	}
	if err := ValidateCount(s.ChannelCount); err != nil {
		return crex.Wrap(ErrInvalidResource, err)
	}
	if err := ValidateTimestamps(s.CreatedAt, s.UpdatedAt); err != nil {
		return crex.Wrap(ErrInvalidResource, err)
	}
	return nil
}

// Complete resource with all its versions and channels.
//
// Provides comprehensive resource information including metadata, versions, and
// channels. The versions and channels lists contain lightweight summary entries
// without full archive details. For complete version information, fetch version
// details. Includes scoping information to identify the resource's location. The
// media type is [MediaTypeResource].
type Resource struct {
	Namespace string            `codec:"namespace"` // Namespace this resource belongs to.
	Name      string            `codec:"name"`      // Resource name.
	Type      string            `codec:"type"`      // Resource type (e.g., "widget", "service").
	Description string          `codec:"description"` // Description.
	Versions  []VersionSummary  `codec:"versions"`  // List of versions (summary form).
	Channels  []ChannelSummary  `codec:"channels"`  // List of channels (summary form).
	CreatedAt int64             `codec:"createdAt"` // When the resource was created.
	UpdatedAt int64             `codec:"updatedAt"` // When the resource was last updated.
}

// Validates the resource.
func (r *Resource) Validate() error {
	if err := ValidateName(r.Namespace); err != nil {
		return crex.Wrap(ErrInvalidResource, err)
	}
	if err := ValidateName(r.Name); err != nil {
		return crex.Wrap(ErrInvalidResource, err)
	}
	if err := ValidateResourceType(r.Type); err != nil {
		return crex.Wrap(ErrInvalidResource, err)
	}
	if err := ValidateTimestamps(r.CreatedAt, r.UpdatedAt); err != nil {
		return crex.Wrap(ErrInvalidResource, err)
	}
	for i := range r.Versions {
		if err := r.Versions[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidResource, err)
		}
	}
	for i := range r.Channels {
		if err := r.Channels[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidResource, err)
		}
	}
	return nil
}

// Collection of resources.
//
// The media type is [MediaTypeResourceList].
type ResourceList struct {

	// List of resources.
	Resources []ResourceSummary `codec:"resources"`
}

// Validates the resource list.
func (l *ResourceList) Validate() error {
	for i := range l.Resources {
		if err := l.Resources[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidResource, err)
		}
	}
	return nil
}
