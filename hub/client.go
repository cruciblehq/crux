package hub

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/cruciblehq/spec/registry"
	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/httpx"
)

// HTTP client for interacting with the registry.
//
// Implements the [registry.Registry] interface over HTTP, providing a remote
// client for registry operations. Handles request serialization, response
// parsing, and error handling according to the Hub API conventions.
type Client struct {
	baseURL    string       // Base URL of the registry.
	httpClient *http.Client // HTTP client used to execute requests.
}

// Creates a new registry client.
//
// The base URL should point to the registry. http.DefaultClient is used if
// httpClient is nil. Returns a new Client instance.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: baseURL, httpClient: httpClient}
}

// Creates a new namespace.
func (c *Client) CreateNamespace(ctx context.Context, info registry.NamespaceInfo) (*registry.Namespace, error) {
	body, err := codec.Encode(info, registry.BodyFormat)
	if err != nil {
		return nil, crex.Wrap(ErrRequestEncode, err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, registry.NamespaceListPath(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderContentType, registry.MediaTypeNamespaceInfo.With(registry.BodyFormat))
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeNamespace.With(registry.BodyFormat))

	var ns registry.Namespace
	if err := c.do(req, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// Retrieves namespace metadata and resource summaries.
func (c *Client) ReadNamespace(ctx context.Context, namespace string) (*registry.Namespace, error) {
	req, err := c.newRequest(ctx, http.MethodGet, registry.NamespacePath(namespace), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeNamespace.With(registry.BodyFormat))

	var ns registry.Namespace
	if err := c.do(req, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// Updates mutable namespace metadata.
func (c *Client) UpdateNamespace(ctx context.Context, namespace string, info registry.NamespaceInfo) (*registry.Namespace, error) {
	body, err := codec.Encode(info, registry.BodyFormat)
	if err != nil {
		return nil, crex.Wrap(ErrRequestEncode, err)
	}

	req, err := c.newRequest(ctx, http.MethodPut, registry.NamespacePath(namespace), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderContentType, registry.MediaTypeNamespaceInfo.With(registry.BodyFormat))
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeNamespace.With(registry.BodyFormat))

	var ns registry.Namespace
	if err := c.do(req, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// Permanently deletes a namespace.
func (c *Client) DeleteNamespace(ctx context.Context, namespace string) error {
	req, err := c.newRequest(ctx, http.MethodDelete, registry.NamespacePath(namespace), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// Lists all namespaces.
func (c *Client) ListNamespaces(ctx context.Context) (*registry.NamespaceList, error) {
	req, err := c.newRequest(ctx, http.MethodGet, registry.NamespaceListPath(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeNamespaceList.With(registry.BodyFormat))

	var list registry.NamespaceList
	if err := c.do(req, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Creates a new resource in the specified namespace.
func (c *Client) CreateResource(ctx context.Context, namespace string, info registry.ResourceInfo) (*registry.Resource, error) {
	body, err := codec.Encode(info, registry.BodyFormat)
	if err != nil {
		return nil, crex.Wrap(ErrRequestEncode, err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, registry.ResourceListPath(namespace), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderContentType, registry.MediaTypeResourceInfo.With(registry.BodyFormat))
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeResource.With(registry.BodyFormat))

	var resource registry.Resource
	if err := c.do(req, &resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

// Retrieves resource metadata with version and channel summaries.
func (c *Client) ReadResource(ctx context.Context, namespace, resource string) (*registry.Resource, error) {
	req, err := c.newRequest(ctx, http.MethodGet, registry.ResourcePath(namespace, resource), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeResource.With(registry.BodyFormat))

	var res registry.Resource
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Updates mutable resource metadata.
func (c *Client) UpdateResource(ctx context.Context, namespace, resource string, info registry.ResourceInfo) (*registry.Resource, error) {
	body, err := codec.Encode(info, registry.BodyFormat)
	if err != nil {
		return nil, crex.Wrap(ErrRequestEncode, err)
	}

	req, err := c.newRequest(ctx, http.MethodPut, registry.ResourcePath(namespace, resource), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderContentType, registry.MediaTypeResourceInfo.With(registry.BodyFormat))
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeResource.With(registry.BodyFormat))

	var res registry.Resource
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Permanently deletes a resource.
func (c *Client) DeleteResource(ctx context.Context, namespace, resource string) error {
	req, err := c.newRequest(ctx, http.MethodDelete, registry.ResourcePath(namespace, resource), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// Lists all resources in a namespace.
func (c *Client) ListResources(ctx context.Context, namespace string) (*registry.ResourceList, error) {
	req, err := c.newRequest(ctx, http.MethodGet, registry.ResourceListPath(namespace), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeResourceList.With(registry.BodyFormat))

	var list registry.ResourceList
	if err := c.do(req, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Creates a new version for a resource.
func (c *Client) CreateVersion(ctx context.Context, namespace, resource string, info registry.VersionInfo) (*registry.Version, error) {
	body, err := codec.Encode(info, registry.BodyFormat)
	if err != nil {
		return nil, crex.Wrap(ErrRequestEncode, err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, registry.VersionListPath(namespace, resource), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderContentType, registry.MediaTypeVersionInfo.With(registry.BodyFormat))
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeVersion.With(registry.BodyFormat))

	var version registry.Version
	if err := c.do(req, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// Retrieves version metadata with archive details.
func (c *Client) ReadVersion(ctx context.Context, namespace, resource, version string) (*registry.Version, error) {
	req, err := c.newRequest(ctx, http.MethodGet, registry.VersionPath(namespace, resource, version), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeVersion.With(registry.BodyFormat))

	var ver registry.Version
	if err := c.do(req, &ver); err != nil {
		return nil, err
	}
	return &ver, nil
}

// Updates mutable version metadata.
func (c *Client) UpdateVersion(ctx context.Context, namespace, resource, version string, info registry.VersionInfo) (*registry.Version, error) {
	body, err := codec.Encode(info, registry.BodyFormat)
	if err != nil {
		return nil, crex.Wrap(ErrRequestEncode, err)
	}

	req, err := c.newRequest(ctx, http.MethodPut, registry.VersionPath(namespace, resource, version), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderContentType, registry.MediaTypeVersionInfo.With(registry.BodyFormat))
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeVersion.With(registry.BodyFormat))

	var ver registry.Version
	if err := c.do(req, &ver); err != nil {
		return nil, err
	}
	return &ver, nil
}

// Permanently deletes a version.
func (c *Client) DeleteVersion(ctx context.Context, namespace, resource, version string) error {
	req, err := c.newRequest(ctx, http.MethodDelete, registry.VersionPath(namespace, resource, version), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// Lists all versions for a resource.
func (c *Client) ListVersions(ctx context.Context, namespace, resource string) (*registry.VersionList, error) {
	req, err := c.newRequest(ctx, http.MethodGet, registry.VersionListPath(namespace, resource), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeVersionList.With(registry.BodyFormat))

	var list registry.VersionList
	if err := c.do(req, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Uploads a version archive.
func (c *Client) UploadArchive(ctx context.Context, namespace, resource, version string, archive io.Reader) (*registry.Version, error) {
	req, err := c.newRequest(ctx, http.MethodPut, registry.ArchivePath(namespace, resource, version), io.NopCloser(archive))
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderContentType, registry.MediaTypeArchive.String())
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeVersion.With(registry.BodyFormat))

	var ver registry.Version
	if err := c.do(req, &ver); err != nil {
		return nil, err
	}
	return &ver, nil
}

// Downloads a version archive.
func (c *Client) DownloadArchive(ctx context.Context, namespace, resource, version string) (io.ReadCloser, error) {
	req, err := c.newRequest(ctx, http.MethodGet, registry.ArchivePath(namespace, resource, version), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeArchive.String())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, crex.SystemError("cannot reach registry", "the request to the registry failed").
			Cause(crex.Wrap(ErrHTTPExecute, err)).
			Err()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		regErr, err := decodeRegistryError(resp)
		if err != nil {
			return nil, crex.SystemErrorf("registry request failed", "the server returned status %d (%s)", resp.StatusCode, resp.Status).
				Cause(ErrHTTPStatus).
				Err()
		}
		return nil, regErr
	}

	return resp.Body, nil
}

// Creates a new channel.
func (c *Client) CreateChannel(ctx context.Context, namespace, resource string, info registry.ChannelInfo) (*registry.Channel, error) {
	body, err := codec.Encode(info, registry.BodyFormat)
	if err != nil {
		return nil, crex.Wrap(ErrRequestEncode, err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, registry.ChannelListPath(namespace, resource), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderContentType, registry.MediaTypeChannelInfo.With(registry.BodyFormat))
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeChannel.With(registry.BodyFormat))

	var channel registry.Channel
	if err := c.do(req, &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

// Retrieves channel metadata with full version details.
func (c *Client) ReadChannel(ctx context.Context, namespace, resource, channel string) (*registry.Channel, error) {
	req, err := c.newRequest(ctx, http.MethodGet, registry.ChannelPath(namespace, resource, channel), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeChannel.With(registry.BodyFormat))

	var ch registry.Channel
	if err := c.do(req, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// Updates a channel's mutable metadata.
func (c *Client) UpdateChannel(ctx context.Context, namespace, resource, channel string, info registry.ChannelInfo) (*registry.Channel, error) {
	body, err := codec.Encode(info, registry.BodyFormat)
	if err != nil {
		return nil, crex.Wrap(ErrRequestEncode, err)
	}

	req, err := c.newRequest(ctx, http.MethodPut, registry.ChannelPath(namespace, resource, channel), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderContentType, registry.MediaTypeChannelInfo.With(registry.BodyFormat))
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeChannel.With(registry.BodyFormat))

	var ch registry.Channel
	if err := c.do(req, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// Permanently deletes a channel.
func (c *Client) DeleteChannel(ctx context.Context, namespace, resource, channel string) error {
	req, err := c.newRequest(ctx, http.MethodDelete, registry.ChannelPath(namespace, resource, channel), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// Lists all channels for a resource.
func (c *Client) ListChannels(ctx context.Context, namespace, resource string) (*registry.ChannelList, error) {
	req, err := c.newRequest(ctx, http.MethodGet, registry.ChannelListPath(namespace, resource), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpx.HeaderAccept, registry.MediaTypeChannelList.With(registry.BodyFormat))

	var list registry.ChannelList
	if err := c.do(req, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Creates an HTTP request with the given method, path, and body.
func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := httpx.NewRequest(ctx, method, c.baseURL, path, nil, body)
	if err != nil {
		return nil, crex.Wrap(ErrHTTPRequest, err)
	}
	return req, nil
}

// Executes an HTTP request and decodes the response body into result.
//
// A response outside the 2xx range carries a [registry.Error], which is
// returned as is. If that error cannot be decoded, a system error is returned
// instead. result may be nil when no body is expected.
func (c *Client) do(req *http.Request, result any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return crex.SystemError("cannot reach registry", "the request to the registry failed").
			Cause(crex.Wrap(ErrHTTPExecute, err)).
			Err()
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		regErr, err := decodeRegistryError(resp)
		if err != nil {
			return crex.SystemErrorf("registry request failed", "the server returned status %d (%s)", resp.StatusCode, resp.Status).
				Cause(ErrHTTPStatus).
				Err()
		}
		return regErr
	}

	if result != nil {
		data, err := io.ReadAll(resp.Body)
		if err == nil {
			err = codec.Decode(data, result, registry.BodyFormat)
		}
		if err != nil {
			return crex.SystemError("cannot read registry response", "the registry returned an unreadable response").
				Cause(crex.Wrap(ErrResponseDecode, err)).
				Err()
		}
	}

	return nil
}

// Decodes the registry error carried by a failed response.
//
// Reads at most [httpx.MaxErrorPayload] bytes of the body.
func decodeRegistryError(resp *http.Response) (*registry.Error, error) {
	data, err := io.ReadAll(io.LimitReader(resp.Body, httpx.MaxErrorPayload))
	if err != nil {
		return nil, err
	}
	var regErr registry.Error
	if err := codec.Decode(data, &regErr, registry.BodyFormat); err != nil {
		return nil, err
	}
	return &regErr, nil
}
