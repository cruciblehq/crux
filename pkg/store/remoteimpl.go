package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/cruciblehq/crux/pkg/reference"
)

const (
	headerAccept         = "Accept"
	headerContentLength  = "Content-Length"
	headerContentType    = "Content-Type"
	headerContentVersion = "Content-Version"
	headerETag           = "ETag"
	headerIfNoneMatch    = "If-None-Match"
	headerUserAgent      = "User-Agent"
)

const (
	defaultAPIVersion           = "v1"
	defaultRemoteTimeout        = 30 * time.Second
	defaultRemoteMaxArchiveSize = 100 * 1024 * 1024
)

// Base URL for the Crucible registry.
var remoteURL = &url.URL{
	Scheme: "https",
	Host:   "registry.crucible.net",
}

// Configuration options for [NewRemote].
type RemoteOptions struct {
	Timeout        time.Duration // Timeout for HTTP requests. Defaults to 30 seconds.
	UserAgent      string        // User-Agent header value.
	MaxArchiveSize int64         // Maximum size for fetched archives. Defaults to 100 MB.
}

// Creates a new client for the official Crucible registry.
//
// The returned [Remote] communicates with the Crucible registry API at
// https://registry.crucible.net. It supports fetching namespace and resource
// metadata as well as resource archives by version or channel.
func NewRemote(options RemoteOptions) Remote {
	timeout := options.Timeout
	if timeout == 0 {
		timeout = defaultRemoteTimeout
	}

	return newRemoteWithHTTPClient(&http.Client{
		Timeout: timeout,
	}, remoteURL, options)
}

// Implementation of [Remote] for the Crucible registry.
type remoteImpl struct {
	client  *http.Client
	baseURL *url.URL
	options RemoteOptions
}

// Creates a new remoteImpl with the given HTTP client and base URL.
//
// This function is intended for testing, allowing injection of a custom
// HTTP client and base URL. Use [NewRemote] for production code.
func newRemoteWithHTTPClient(httpClient *http.Client, baseURL *url.URL, options RemoteOptions) *remoteImpl {
	maxArchiveSize := options.MaxArchiveSize
	if maxArchiveSize == 0 {
		maxArchiveSize = defaultRemoteMaxArchiveSize
	}

	return &remoteImpl{
		client:  httpClient,
		baseURL: baseURL,
		options: RemoteOptions{
			UserAgent:      options.UserAgent,
			MaxArchiveSize: maxArchiveSize,
		},
	}
}

// Returns namespace metadata and resource list.
//
// Sends a GET request to /v1/{namespace}. If the request includes an ETag
// from a previous response and the server returns 304 Not Modified, the
// response will have Info set to nil.
func (r *remoteImpl) Namespace(ctx context.Context, req *NamespaceRequest) (*NamespaceResponse, error) {
	u := r.baseURL.JoinPath(defaultAPIVersion, req.Namespace)

	resp, err := r.get(ctx, u.String(), MediaTypeNamespace, req.ETag)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return &NamespaceResponse{ETag: req.ETag}, nil
	}

	if err := r.checkError(resp); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status %d", ErrFetchFailed, resp.StatusCode)
	}

	if err := r.checkContentType(resp, MediaTypeNamespace); err != nil {
		return nil, err
	}

	var info NamespaceInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}

	return &NamespaceResponse{
		Info: &info,
		ETag: resp.Header.Get(headerETag),
	}, nil
}

// Returns resource metadata, available versions, and channels.
//
// Sends a GET request to /v1/{namespace}/{name}. If the request includes
// an ETag from a previous response and the server returns 304 Not Modified,
// the response will have Info set to nil.
func (r *remoteImpl) Resource(ctx context.Context, req *ResourceRequest) (*ResourceResponse, error) {
	u := r.baseURL.JoinPath(defaultAPIVersion, req.Identifier.Namespace(), req.Identifier.Name())

	resp, err := r.get(ctx, u.String(), MediaTypeResource, req.ETag)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return &ResourceResponse{ETag: req.ETag}, nil
	}

	if err := r.checkError(resp); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status %d", ErrFetchFailed, resp.StatusCode)
	}

	if err := r.checkContentType(resp, MediaTypeResource); err != nil {
		return nil, err
	}

	var info ResourceInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}

	return &ResourceResponse{
		Info: &info,
		ETag: resp.Header.Get(headerETag),
	}, nil
}

// Fetches a resource archive by version.
//
// Sends a GET request to /v1/{namespace}/{name}/{version}. If the request
// includes an ETag from a previous response and the server returns 304 Not
// Modified, the response will have Data set to nil.
//
// The archive size is limited to [RemoteOptions.MaxArchiveSize]. If the
// archive exceeds this limit, an error is returned.
func (r *remoteImpl) Fetch(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
	u := r.baseURL.JoinPath(defaultAPIVersion, req.Identifier.Namespace(), req.Identifier.Name(), req.Version.String())

	resp, err := r.get(ctx, u.String(), MediaTypeArchive, req.ETag)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return &FetchResponse{ETag: req.ETag}, nil
	}

	if err := r.checkError(resp); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status %d", ErrFetchFailed, resp.StatusCode)
	}

	if err := r.checkContentType(resp, MediaTypeArchive); err != nil {
		return nil, err
	}

	data, err := r.readArchive(resp)
	if err != nil {
		return nil, err
	}

	return &FetchResponse{
		Data: data,
		ETag: resp.Header.Get(headerETag),
	}, nil
}

// Fetches a resource archive by channel.
//
// Sends a GET request to /v1/{namespace}/{name}/:{channel}. The server
// resolves the channel to the latest version and returns the archive.
//
// If the request includes an ETag from a previous response and the server
// returns 304 Not Modified, the response will have Data set to nil.
//
// The archive size is limited to [RemoteOptions.MaxArchiveSize]. If the
// archive exceeds this limit, an error is returned.
func (r *remoteImpl) Consume(ctx context.Context, req *ConsumeRequest) (*ConsumeResponse, error) {
	u := r.baseURL.JoinPath(defaultAPIVersion, req.Identifier.Namespace(), req.Identifier.Name(), ":"+req.Channel)

	resp, err := r.get(ctx, u.String(), MediaTypeArchive, req.ETag)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	version, err := r.parseVersionHeader(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotModified {
		return &ConsumeResponse{
			Version: *version,
			ETag:    req.ETag,
		}, nil
	}

	if err := r.checkError(resp); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status %d", ErrFetchFailed, resp.StatusCode)
	}

	if err := r.checkContentType(resp, MediaTypeArchive); err != nil {
		return nil, err
	}

	data, err := r.readArchive(resp)
	if err != nil {
		return nil, err
	}

	return &ConsumeResponse{
		Version: *version,
		Data:    data,
		ETag:    resp.Header.Get(headerETag),
	}, nil
}

// Sends a GET request with Accept header and optional ETag for conditional requests.
func (r *remoteImpl) get(ctx context.Context, url, accept, etag string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}

	req.Header.Set(headerAccept, accept)

	if r.options.UserAgent != "" {
		req.Header.Set(headerUserAgent, r.options.UserAgent)
	}

	if etag != "" {
		req.Header.Set(headerIfNoneMatch, etag)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}

	return resp, nil
}

// Parses the Content-Version header from the response.
func (r *remoteImpl) parseVersionHeader(resp *http.Response) (*reference.Version, error) {
	versionStr := resp.Header.Get(headerContentVersion)
	if versionStr == "" {
		return nil, fmt.Errorf("%w: missing %s header", ErrFetchFailed, headerContentVersion)
	}

	version, err := reference.ParseVersion(versionStr)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid %s header: %v", ErrFetchFailed, headerContentVersion, err)
	}

	return version, nil
}

// Checks for error responses and parses RegistryError.
func (r *remoteImpl) checkError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	contentType := resp.Header.Get(headerContentType)
	if contentType == "" {
		return nil
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil
	}

	if mediaType != MediaTypeError {
		return nil
	}

	var regErr RegistryError
	if err := json.NewDecoder(resp.Body).Decode(&regErr); err != nil {
		return fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}

	return &regErr
}

// Validates the response Content-Type header.
func (r *remoteImpl) checkContentType(resp *http.Response, expected string) error {
	contentType := resp.Header.Get(headerContentType)
	if contentType == "" {
		return fmt.Errorf("%w: missing content type", ErrFetchFailed)
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("%w: invalid content type: %v", ErrFetchFailed, err)
	}

	if mediaType != expected {
		return fmt.Errorf("%w: unexpected content type %s", ErrFetchFailed, mediaType)
	}

	return nil
}

// Reads archive data from the response body with size limit.
//
// Checks Content-Length header first if available. Then reads the body with a
// limit of MaxArchiveSize + 1 to detect oversized archives.
func (r *remoteImpl) readArchive(resp *http.Response) ([]byte, error) {
	if resp.ContentLength > r.options.MaxArchiveSize {
		return nil, fmt.Errorf("%w: archive too large (%d bytes)", ErrFetchFailed, resp.ContentLength)
	}

	limitedReader := io.LimitReader(resp.Body, r.options.MaxArchiveSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}

	if int64(len(data)) > r.options.MaxArchiveSize {
		return nil, fmt.Errorf("%w: archive too large (exceeds %d bytes)", ErrFetchFailed, r.options.MaxArchiveSize)
	}

	return data, nil
}
