package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/cruciblehq/crux/crex"
)

const (
	headerContentType = "Content-Type" // Content-Type HTTP header.
	headerAccept      = "Accept"       // Accept HTTP header.
)

// HTTP client for interacting with the Crucible Hub registry.
//
// Implements the Registry interface over HTTP, providing a remote client for
// registry operations. Handles request serialization, response parsing, and
// error handling according to the Hub API conventions.
type Client struct {
	baseURL    string       // Base URL of the Hub registry.
	httpClient *http.Client // HTTP client used to execute requests.
}

// Creates a new Hub client.
//
// The base URL should point to the Hub registry. If httpClient is nil,
// http.DefaultClient is used.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// Creates a new namespace.
func (c *Client) CreateNamespace(ctx context.Context, info NamespaceInfo) (*Namespace, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, namespacesPath(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerContentType, MediaTypeNamespaceInfo.JSON())
	req.Header.Set(headerAccept, MediaTypeNamespace.JSON())

	var ns Namespace
	if err := c.do(req, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// Retrieves namespace metadata and resource summaries.
func (c *Client) ReadNamespace(ctx context.Context, namespace string) (*Namespace, error) {
	req, err := c.newRequest(ctx, http.MethodGet, namespacePath(namespace), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerAccept, MediaTypeNamespace.JSON())

	var ns Namespace
	if err := c.do(req, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// Updates mutable namespace metadata.
func (c *Client) UpdateNamespace(ctx context.Context, namespace string, info NamespaceInfo) (*Namespace, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	req, err := c.newRequest(ctx, http.MethodPut, namespacePath(namespace), bytes.NewReader(body))
	req.Header.Set(headerAccept, MediaTypeNamespace.JSON())

	var ns Namespace
	if err := c.do(req, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// Permanently deletes a namespace.
func (c *Client) DeleteNamespace(ctx context.Context, namespace string) error {
	req, err := c.newRequest(ctx, http.MethodDelete, namespacePath(namespace), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// Lists all namespaces.
func (c *Client) ListNamespaces(ctx context.Context) (*NamespaceList, error) {
	req, err := c.newRequest(ctx, http.MethodGet, namespacesPath(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerAccept, MediaTypeNamespaceList.JSON())

	var list NamespaceList
	if err := c.do(req, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Creates a new resource in the specified namespace.
func (c *Client) CreateResource(ctx context.Context, namespace string, info ResourceInfo) (*Resource, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, resourcesPath(namespace), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerContentType, MediaTypeResourceInfo.JSON())
	req.Header.Set(headerAccept, MediaTypeResource.JSON())

	var resource Resource
	if err := c.do(req, &resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

// Retrieves resource metadata with version and channel summaries.
func (c *Client) ReadResource(ctx context.Context, namespace, resource string) (*Resource, error) {
	req, err := c.newRequest(ctx, http.MethodGet, resourcePath(namespace, resource), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerAccept, MediaTypeResource.JSON())

	var res Resource
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Updates mutable resource metadata.
func (c *Client) UpdateResource(ctx context.Context, namespace, resource string, info ResourceInfo) (*Resource, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	req, err := c.newRequest(ctx, http.MethodPut, resourcePath(namespace, resource), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerContentType, MediaTypeResourceInfo.JSON())
	req.Header.Set(headerAccept, MediaTypeResource.JSON())

	var res Resource
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Permanently deletes a resource.
func (c *Client) DeleteResource(ctx context.Context, namespace, resource string) error {
	req, err := c.newRequest(ctx, http.MethodDelete, resourcePath(namespace, resource), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// Lists all resources in a namespace.
func (c *Client) ListResources(ctx context.Context, namespace string) (*ResourceList, error) {
	req, err := c.newRequest(ctx, http.MethodGet, resourcesPath(namespace), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerAccept, MediaTypeResourceList.JSON())

	var list ResourceList
	if err := c.do(req, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Creates a new version for a resource.
func (c *Client) CreateVersion(ctx context.Context, namespace, resource string, info VersionInfo) (*Version, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, versionsPath(namespace, resource), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerContentType, MediaTypeVersionInfo.JSON())
	req.Header.Set(headerAccept, MediaTypeVersion.JSON())

	var version Version
	if err := c.do(req, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// Retrieves version metadata with archive details.
func (c *Client) ReadVersion(ctx context.Context, namespace, resource, version string) (*Version, error) {
	req, err := c.newRequest(ctx, http.MethodGet, versionPath(namespace, resource, version), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerAccept, MediaTypeVersion.JSON())

	var ver Version
	if err := c.do(req, &ver); err != nil {
		return nil, err
	}
	return &ver, nil
}

// Updates mutable version metadata.
func (c *Client) UpdateVersion(ctx context.Context, namespace, resource, version string, info VersionInfo) (*Version, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	req, err := c.newRequest(ctx, http.MethodPut, versionPath(namespace, resource, version), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerContentType, MediaTypeVersionInfo.JSON())
	req.Header.Set(headerAccept, MediaTypeVersion.JSON())

	var ver Version
	if err := c.do(req, &ver); err != nil {
		return nil, err
	}
	return &ver, nil
}

// Permanently deletes a version.
func (c *Client) DeleteVersion(ctx context.Context, namespace, resource, version string) error {
	req, err := c.newRequest(ctx, http.MethodDelete, versionPath(namespace, resource, version), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// Lists all versions for a resource.
func (c *Client) ListVersions(ctx context.Context, namespace, resource string) (*VersionList, error) {
	req, err := c.newRequest(ctx, http.MethodGet, versionsPath(namespace, resource), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerAccept, MediaTypeVersionList.JSON())

	var list VersionList
	if err := c.do(req, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Uploads a version archive.
func (c *Client) UploadArchive(ctx context.Context, namespace, resource, version string, archive io.Reader) (*Version, error) {
	req, err := c.newRequest(ctx, http.MethodPut, archivePath(namespace, resource, version), archive)
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerContentType, string(MediaTypeArchive))
	req.Header.Set(headerAccept, MediaTypeVersion.JSON())

	var ver Version
	if err := c.do(req, &ver); err != nil {
		return nil, err
	}
	return &ver, nil
}

// Downloads a version archive.
func (c *Client) DownloadArchive(ctx context.Context, namespace, resource, version string) (io.ReadCloser, error) {
	req, err := c.newRequest(ctx, http.MethodGet, archivePath(namespace, resource, version), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerAccept, string(MediaTypeArchive))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, crex.Wrap(ErrHTTPExecute, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		var regErr Error
		if err := json.NewDecoder(resp.Body).Decode(&regErr); err != nil {
			return nil, crex.Wrapf(ErrHTTPStatus, "HTTP %d: %s", resp.StatusCode, resp.Status)
		}
		return nil, &regErr
	}

	return resp.Body, nil
}

// Creates a new channel.
func (c *Client) CreateChannel(ctx context.Context, namespace, resource string, info ChannelInfo) (*Channel, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, channelsPath(namespace, resource), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerContentType, MediaTypeChannelInfo.JSON())
	req.Header.Set(headerAccept, MediaTypeChannel.JSON())

	var channel Channel
	if err := c.do(req, &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

// Retrieves channel metadata with full version details.
func (c *Client) ReadChannel(ctx context.Context, namespace, resource, channel string) (*Channel, error) {
	req, err := c.newRequest(ctx, http.MethodGet, channelPath(namespace, resource, channel), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerAccept, MediaTypeChannel.JSON())

	var ch Channel
	if err := c.do(req, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// Updates a channel's mutable metadata.
func (c *Client) UpdateChannel(ctx context.Context, namespace, resource, channel string, info ChannelInfo) (*Channel, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	req, err := c.newRequest(ctx, http.MethodPut, channelPath(namespace, resource, channel), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerContentType, MediaTypeChannelInfo.JSON())
	req.Header.Set(headerAccept, MediaTypeChannel.JSON())

	var ch Channel
	if err := c.do(req, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// Permanently deletes a channel.
func (c *Client) DeleteChannel(ctx context.Context, namespace, resource, channel string) error {
	req, err := c.newRequest(ctx, http.MethodDelete, channelPath(namespace, resource, channel), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// Lists all channels for a resource.
func (c *Client) ListChannels(ctx context.Context, namespace, resource string) (*ChannelList, error) {
	req, err := c.newRequest(ctx, http.MethodGet, channelsPath(namespace, resource), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerAccept, MediaTypeChannelList.JSON())

	var list ChannelList
	if err := c.do(req, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Creates an HTTP request with the given method, path, and body.
func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, crex.Wrap(ErrBaseURL, err)
	}
	u.Path = path

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, crex.Wrap(ErrHTTPRequest, err)
	}
	return req, nil
}

// Executes an HTTP request and decodes the JSON response.
func (c *Client) do(req *http.Request, result interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return crex.Wrap(ErrHTTPExecute, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var regErr Error
		if err := json.NewDecoder(resp.Body).Decode(&regErr); err != nil {
			return crex.Wrapf(ErrHTTPStatus, "HTTP %d: %s", resp.StatusCode, resp.Status)
		}
		return &regErr
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return crex.Wrap(ErrResponseDecode, err)
		}
	}

	return nil
}
