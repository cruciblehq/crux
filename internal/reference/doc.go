// Package reference defines the structure and parsing logic for Crucible resource references.
//
// A resource reference is a symbolic string that encapsulates the information
// needed to fetch and verify a resource. Its general shape is an optional
// resource type, an optional [scheme://]registry, a hierarchical path, a
// version constraint or a channel, and an optional digest, written for
// example as "widget official/my-widget ^1.2.0 sha256:3a7bd3e2360a3d80c1...".
// Scheme, registry, and path together form the resource location, and when
// scheme or registry are omitted they default to the values configured in
// [Options]. The path itself uses Crucible's namespace/name convention, where
// the namespace groups related resources and the name identifies one within
// that namespace; the namespace may be omitted only when the registry is also
// omitted, in which case the configured default namespace applies. As a
// result, the references "my-widget", "official/my-widget",
// "registry.crucible.net/official/my-widget", and
// "http://registry.crucible.net/official/my-widget" all denote the same
// resource when the defaults match.
//
// Versioning follows semantic versioning with a small number of deliberate
// deviations. The version segment accepts the operators that semver libraries
// usually support: coercion of partial versions, OR with || and implicit AND
// by whitespace, the comparison operators =, !=, >, <, >=, and <=, hyphen
// ranges, x-style wildcards, tilde patch ranges, and caret minor ranges.
// Crucible diverges from semver by requiring every version range to have an
// explicit upper bound, since future major versions introduce breaking changes
// that cannot be anticipated; ">=2.0.0 <3.0.0" and ">1.5.0 <2.0.0" are valid,
// while ">=2.0.0" and ">1.5.0" are not. Operators that imply an upper bound
// such as ^1.2.3, ~1.2.3, 1.2.x, and 1.2.3 - 2.0.0 are allowed because the
// bound is unambiguous. Crucible also rejects the bare asterisk wildcard
// because it undermines the stability guarantees that versioning is meant to
// provide.
//
// Pre-releases receive special treatment. Standard semver compares pre-release
// identifiers lexically in ASCII order, which makes "BETA" sort below "alpha"
// and produces ordering that is unhelpful for Crucible's usage. More
// importantly, pre-releases are inherently unstable and unsuitable for
// Crucible's dynamic composition model, so they are prohibited in version
// constraints except in narrow scenarios where authorised users opt in to
// pre-release access. To cover the common need for tracking unstable streams
// without pre-release identifiers, Crucible introduces channels: named release
// tracks such as "stable", "beta", or "alpha" that act as mutable pointers to
// the latest version on that track. Channels require explicit opt-in and
// authorisation, are unavailable by default, and cannot appear in published
// resources, which keeps them confined to development and testing.
//
// A channel is specified in place of a version with a colon prefix, for
// example "my-widget :stable" or "official/my-widget :beta", and when present
// no other version constraint applies. When a resource is prepared for
// deployment its dependencies are resolved to concrete versions and the
// references are frozen by appending a digest. The original constraints are
// preserved alongside the digest for auditing, but the digest is what
// downstream tooling uses to fetch and verify content. A digest is a
// cryptographic hash of the resource content, written for example as
// "sha256:3a7bd3e2360a3d80c1...", and turns the reference into a content-
// addressed pointer that always denotes the same bytes regardless of any
// changes to the symbolic components.
//
// References also carry a resource type. The type is not represented in the
// reference string itself; it is supplied contextually when a reference is
// parsed in a position where a particular type is expected. When the expected
// type is ambiguous, callers may supply it explicitly as the leading token,
// for example "widget my-widget :stable sha256:3a7bd3e2360a3d80c1...".
//
// Parsing a reference and inspecting its components:
//
//	ref, err := reference.Parse("official/my-widget ^1.2.0", opts)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(ref.Path, ref.Version)
//	fmt.Println(ref.String())
package reference
