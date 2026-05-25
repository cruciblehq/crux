package recipe

import "github.com/cruciblehq/crux/compute"

// Low-level backend for building OCI images.
//
// Alias for [compute.ImageBuilder]; defined here so recipe callers do not
// need to import compute directly.
type Backend = compute.ImageBuilder
