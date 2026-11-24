package hub

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrRequestEncode       = crex.New("failed to encode request body")
	ErrHTTPRequest         = crex.New("failed to create HTTP request")
	ErrHTTPExecute         = crex.New("failed to execute HTTP request")
	ErrHTTPStatus          = crex.New("unexpected HTTP status")
	ErrResponseDecode      = crex.New("failed to decode response body")
	ErrTypeMismatch        = crex.New("resource type mismatch")
	ErrNoVersions          = crex.New("no versions found")
	ErrNoMatchingVersion   = crex.New("no matching version found")
	ErrMissingOption       = crex.New("missing required option")
	ErrMissingRegistry     = crex.New("default registry is required")
	ErrMissingNamespace    = crex.New("default namespace is required")
	ErrCacheOperation      = crex.New("cache operation failed")
	ErrFileSystemOperation = crex.New("file system operation failed")
	ErrNoArchive           = crex.New("version has no uploaded archive")
	ErrResolveVersion      = crex.New("failed to resolve version")
	ErrDownload            = crex.New("failed to download archive")
	ErrInvalidIdentifier   = crex.New("invalid resource identifier")
	ErrNamespaceNotFound   = crex.New("namespace not found")
	ErrVersionPublic       = crex.New("version is public and cannot change")
	ErrVersionSuperseded   = crex.New("a higher public version has the same major and minor")
	ErrRegistryOperation   = crex.New("registry operation failed")
)
