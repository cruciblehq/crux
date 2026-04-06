package registry

import "github.com/cruciblehq/crex"

// Mutable properties of a namespace for creation or update.
//
// Used as the request body for namespace creation and update operations. The
// name is the unique identifier for the namespace and cannot be changed once
// created. For update requests, Name must match the URL path parameter or
// update context. Contains only user-modifiable fields; system-managed fields
// are set by the server and appear only in response types. The media type is
// [MediaTypeNamespaceInfo].
type NamespaceInfo struct {
	Name        string `codec:"name"`        // Namespace name.
	Description string `codec:"description"` // Description.
}

// Validates the namespace info.
//
// The name must conform to the registry naming rules (see [ValidateName]).
func (info *NamespaceInfo) Validate() error {
	if err := ValidateName(info.Name); err != nil {
		return crex.Wrap(ErrInvalidNamespace, err)
	}
	return nil
}

// Lightweight namespace representation for listings.
//
// Provides namespace metadata without nested resource details. Used in list
// responses to keep payloads compact. Includes read-only fields like creation
// timestamps and statistics that are not present in [NamespaceInfo].
type NamespaceSummary struct {
	Name          string `codec:"name"`          // Namespace name.
	Description   string `codec:"description"`   // Description.
	ResourceCount int    `codec:"resourceCount"` // Number of resources in this namespace.
	CreatedAt     int64  `codec:"createdAt"`     // When the namespace was created.
	UpdatedAt     int64  `codec:"updatedAt"`     // When the namespace was last updated.
}

// Validates the namespace summary.
func (s *NamespaceSummary) Validate() error {
	if err := ValidateName(s.Name); err != nil {
		return crex.Wrap(ErrInvalidNamespace, err)
	}
	if err := ValidateCount(s.ResourceCount); err != nil {
		return crex.Wrap(ErrInvalidNamespace, err)
	}
	if err := ValidateTimestamps(s.CreatedAt, s.UpdatedAt); err != nil {
		return crex.Wrap(ErrInvalidNamespace, err)
	}
	return nil
}

// Complete namespace with its resource listings.
//
// Serves as the organizational unit for grouping resources. The resources list
// contains lightweight [ResourceSummary] entries without full version and channel
// details. For complete resource information, fetch individual resources. The
// media type is [MediaTypeNamespace].
type Namespace struct {
	Name        string            `codec:"name"`        // Namespace name.
	Description string            `codec:"description"` // Description.
	Resources   []ResourceSummary `codec:"resources"`   // List of resources (summary form).
	CreatedAt   int64             `codec:"createdAt"`   // When the namespace was created.
	UpdatedAt   int64             `codec:"updatedAt"`   // When the namespace was last updated.
}

// Validates the namespace.
func (ns *Namespace) Validate() error {
	if err := ValidateName(ns.Name); err != nil {
		return crex.Wrap(ErrInvalidNamespace, err)
	}
	if err := ValidateTimestamps(ns.CreatedAt, ns.UpdatedAt); err != nil {
		return crex.Wrap(ErrInvalidNamespace, err)
	}
	for i := range ns.Resources {
		if err := ns.Resources[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidNamespace, err)
		}
	}
	return nil
}

// Collection of namespaces.
//
// Namespaces may be empty if the registry contains no namespaces. The media
// type is [MediaTypeNamespaceList].
type NamespaceList struct {

	// List of namespaces.
	Namespaces []NamespaceSummary `codec:"namespaces"`
}

// Validates the namespace list.
func (l *NamespaceList) Validate() error {
	for i := range l.Namespaces {
		if err := l.Namespaces[i].Validate(); err != nil {
			return crex.Wrap(ErrInvalidNamespace, err)
		}
	}
	return nil
}
