// Package hub provides Crux integration with the Crucible Hub registry.
//
// The shared wire contracts and validation rules for registry entities live in
// github.com/cruciblehq/spec/registry. This builds on top of those types
// with Crux-specific behavior.
//
// Source is the high-level entry point. It applies default registry and
// namespace values while parsing references, resolves concrete versions,
// performs pull and push operations, and integrates with the local cache.
// PullResult reports resolved artifact metadata and the extraction directory.
//
// Client is the HTTP adapter used by this package to talk to the Hub API. Its
// method signatures use the shared types from github.com/cruciblehq/spec/registry.
//
// Example:
//
//	src, err := hub.NewSource("http://hub.cruciblehq.xyz:8080", "acme")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	ref, err := src.Parse("service", "acme/payments ^1.2.0")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	result, err := src.Pull(ctx, ref)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	_ = result.Extracted
package hub
