// Package pull provides functionality for pulling resources from remote registries.
//
// Pull operations download resource archives from a remote registry and store
// them in the local cache. If the resource is already cached, no download
// occurs. The cache is used to avoid redundant downloads and to provide
// offline access to previously fetched resources.
package pull
