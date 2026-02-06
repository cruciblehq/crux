// Package cache provides a local file-locked registry cache for resources.
//
// The cache stores downloaded resource archives locally, avoiding redundant
// downloads from remote registries. It uses the shared SQLRegistry
// implementation from protocol/pkg/registry for storage. All operations are
// protected by file locks to allow safe concurrent access from multiple
// processes (e.g., crux CLI and the local development server).
package cache
