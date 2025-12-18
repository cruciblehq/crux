// Package reference defines the structure and parsing logic for Crucible
// resource references.
//
// Resources are identified and located using resource references, which are
// symbolic strings that encapsulate the necessary information to fetch and
// verify the resource.
//
// A reference has the general format:
//
//	[<type>] [[scheme://]registry/]<path> (<version> | <channel>) [<digest>]
//
// Scheme, registry (authority), and path together form the resource identifier
// (or location), which specifies where to locate the resource. By default, the
// scheme is [DefaultScheme] and the registry [DefaultRegistry], corresponding
// to Crucible's default registry.
//
// Crucible's registry uses a hierarchical path structure to organize resources,
// consisting of namespace and name segments expressed as namespace/name. The
// namespace groups related resources, while the name identifies a specific
// resource within that namespace. The namespace can be omitted, in which case
// it defaults to [DefaultNamespace] ("official"), but only if the registry is
// is not specified. Other registries may use different conventions.
//
// For example, the following references are equivalent:
//
//   - my-widget
//   - official/my-widget
//   - registry.crucible.net/official/my-widget
//   - http://registry.crucible.net/official/my-widget
//
// In the Crucible ecosystem, versioning adheres strictly to semantic versioning
// principles, with a few exceptions. The version segment allows most of the same
// operators usually supported by semantic versioning libraries, specifically:
//
//   - Coercion (1.2 -> 1.2.0, v1.2.3 -> 1.2.3)
//   - OR (1.2 || 1.3) and AND (space-separated) operators
//   - Comparison operators (=, !=, >, <, >=, <=)
//   - Hyphen ranges (1.2.3 - 2.0.0)
//   - Wildcards (1.2.x, 1.x)
//   - Patch-range comparisons (~1.2.3)
//   - Caret-range comparisons (^1.2.3)
//
// The first exception involves unbounded ranges. Often software projects declare
// support for versions using open-ended constraints (e.g., >=2.0.0 or >1.5.0).
// However, this breaks compatibility guarantees, since future major versions
// introduce breaking changes that cannot be anticipated. Therefore, Crucible
// requires all version ranges to have explicit upper bounds. For example:
//
//   - >=2.0.0 <3.0.0 (valid)
//   - >1.5.0 <2.0.0 (valid)
//   - >=2.0.0 (invalid, no upper bound)
//   - >1.5.0 (invalid, no upper bound)
//
// Operators that implicitly define upper bounds are allowed:
//
//   - ^1.2.3 (implies <2.0.0)
//   - ~1.2.3 (implies <1.3.0)
//   - 1.2.x (implies >=1.2.0 <1.3.0)
//   - 1.2.3 - 2.0.0 (explicit range)
//
// Another exception involves wildcards. The asterisk (*) operator is often used
// to denote "any version". In Crucible, this operator is not supported, since
// it undermines the purpose of versioning, which is to provide stability and
// compatibility guarantees. Instead, users should specify explicit version
// constraints that reflect their compatibility requirements.
//
// Pre-releases are also handled differently. The semver specification dictates
// that pre-release identifiers with letters or hyphens are case-sensitive and
// compared lexically in ASCII sort order, entailing "BETA" representing a lower
// version than "alpha". This is contrary to Crucible's intended usage.
//
// Furthermore, semver also specifies that pre-releases are unstable and may not
// satisfy compatibility requirements, making them unsuitable for Crucible's
// dynamic composition use cases. There are exceptions to this, such as when
// preparing for an upcoming release, in which case Crucible enables specific
// users to reference pre-release versions (e.g., those given explicit access).
// However, in general, pre-releases are prohibited in version constraints.
//
// Crucible fixes this problem by introducing channels, which are named release
// tracks (e.g., "stable", "beta", "alpha"). Channels represent mutable version
// streams, allowing users to track different stability levels without dealing
// with pre-release versioning complexities. Channels require explicit opt-in
// and authorization and are not used by default. At the same time, a Crucible
// resource cannot be published with channel dependencies and must use standard
// versioning. Channels are only available for development and testing purposes.
//
// A channel can be specified in place of a version, using a colon prefix (e.g.,
// :stable). When a channel is specified, the latest version in that channel is
// used and no other version constraints apply. For example:
//
//   - my-widget :stable
//   - official/my-widget :beta
//   - registry.crucible.net/official/my-widget :alpha
//   - http://registry.crucible.net/official/my-widget :stable
//
// When a resource is prepared for deployment its dependencies are resolved to
// specific versions and all references are frozen by including a digest. The
// original version constraints are preserved for auditing purposes, but the
// content-addressed reference is used for fetching and verification.
//
// The digest segment provides a cryptographic hash (e.g., SHA-256) of the
// resource content, ensuring immutability and integrity. When a digest is
// included, the reference is considered "frozen" and always refers to the
// exact same content, regardless of any changes to the symbolic components
// of the reference. For example:
//
//   - my-widget >1.2.0 <3.0.0 sha256:3a7bd3e2360a3d80c1...
//   - official/my-widget ^1.2.0 sha256:4b825dc642cb6eb9...
//   - registry.crucible.net/official/my-widget ~1 sha256:5d41402abc4b2a76...
//   - http://registry.crucible.net/official/my-widget =2.0.1 sha256:6f5902ac237024bdd0...
//
// References also include a resource type, although it is not represented in
// in the reference string, instead being provided contextually when parsing
// references. For example, when parsing a reference where a widget is expected,
// the resource type is implicitly "widget". In cases where the resource type
// might be ambiguous, it will have to be provided explicitly:
//
//   - widget my-widget :stable sha256:3a7bd3e2360a3d80c1...
//
// References are parsed using [Parse], which validates the format and extract
// the individual components. The [Reference] struct provides access to the
// parsed components and a [String] method to obtain the canonical string
// representation of the reference.
package reference
