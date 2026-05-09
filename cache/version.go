package cache

// Local representation of a cached resource version.
//
// Tracks what is stored on disk for a given namespace/resource/version triple.
// This is a storage type — it reflects what the cache holds, not the registry
// wire format.
type Version struct {
	Namespace string  `json:"namespace"` // Namespace this version belongs to.
	Resource  string  `json:"resource"`  // Resource this version belongs to.
	String    string  `json:"string"`    // Version string (e.g., "1.0.0").
	Archive   *string `json:"archive"`   // Archive filename within the version directory.
	Size      *int64  `json:"size"`      // Archive size in bytes.
	Digest    *string `json:"digest"`    // Archive content digest (e.g., "sha256:abc...").
	CreatedAt int64   `json:"createdAt"` // When the entry was first stored.
	UpdatedAt int64   `json:"updatedAt"` // When the entry was last updated.
}
