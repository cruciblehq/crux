// Package source coordinates pulling and pushing Crucible resources between a
// registry and the local cache.
//
// A [Source] holds the default registry URL and namespace applied to reference
// strings that omit those components. [Source.Parse] resolves a reference
// string into a [reference.Reference], filling in the defaults. [Source.Pull]
// downloads and extracts a resource, consulting the local cache first and
// verifying the artifact digest before use. [Source.Push] uploads a package,
// creating the resource and version in the registry when they do not yet
// exist.
//
// A [PullResult] reports the local extraction directory along with the digest
// and version metadata of the pulled artifact.
package source
