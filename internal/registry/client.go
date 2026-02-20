package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/cruciblehq/crex"
	specregistry "github.com/cruciblehq/spec/registry"
)

// HTTP client for interacting with the Crucible Hub registry.
//
// Implements the Registry interface over HTTP, providing a remote client for
// registry operations. Handles request serialization, response parsing, and
// error handling according to the Hub API conventions.
type Client struct {
	baseURL    string
	httpClient *http.Client
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
func (c *Client) CreateNamespace(ctx context.Context, info specregistry.NamespaceInfo) (*specregistry.Namespace, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	req, err := c.newRequest(ctx, "POST", "/namespaces", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(specregistry.MediaTypeNamespaceInfo)+"+json")
	req.Header.Set("Accept", string(specregistry.MediaTypeNamespace)+"+json")

	var ns specregistry.Namespace
	if err := c.do(req, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// Retrieves namespace metadata and resource summaries.
func (c *Client) ReadNamespace(ctx context.Context, namespace string) (*specregistry.Namespace, error) {
	path, _ := url.JoinPath("/namespaces", namespace)
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(specregistry.MediaTypeNamespace)+"+json")

	var ns specregistry.Namespace
	if err := c.do(req, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// Updates mutable namespace metadata.
func (c *Client) UpdateNamespace(ctx context.Context, namespace string, info specregistry.NamespaceInfo) (*specregistry.Namespace, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	path, _ := url.JoinPath("/namespaces", namespace)
	req, err := c.newRequest(ctx, "PUT", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(specregistry.MediaTypeNamespaceInfo)+"+json")
	req.Header.Set("Accept", string(specregistry.MediaTypeNamespace)+"+json")

	var ns specregistry.Namespace
	if err := c.do(req, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// Permanently deletes a namespace.
func (c *Client) DeleteNamespace(ctx context.Context, namespace string) error {
	path, _ := url.JoinPath("/namespaces", namespace)
	req, err := c.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// Lists all namespaces.
func (c *Client) ListNamespaces(ctx context.Context) (*specregistry.NamespaceList, error) {
	req, err := c.newRequest(ctx, "GET", "/namespaces", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(specregistry.MediaTypeNamespaceList)+"+json")

	var list specregistry.NamespaceList
	if err := c.do(req, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Creates a new resource in the specified namespace.
func (c *Client) CreateResource(ctx context.Context, namespace string, info specregistry.ResourceInfo) (*specregistry.Resource, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	path, _ := url.JoinPath("/namespaces", namespace, "resources")
	req, err := c.newRequest(ctx, "POST", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(specregistry.MediaTypeResourceInfo)+"+json")
	req.Header.Set("Accept", string(specregistry.MediaTypeResource)+"+json")

	var resource specregistry.Resource
	if err := c.do(req, &resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

// Retrieves resource metadata with version and channel summaries.
func (c *Client) ReadResource(ctx context.Context, namespace, resource string) (*specregistry.Resource, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource)
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(specregistry.MediaTypeResource)+"+json")

	var res specregistry.Resource
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Updates mutable resource metadata.
func (c *Client) UpdateResource(ctx context.Context, namespace, resource string, info specregistry.ResourceInfo) (*specregistry.Resource, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource)
	req, err := c.newRequest(ctx, "PUT", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(specregistry.MediaTypeResourceInfo)+"+json")
	req.Header.Set("Accept", string(specregistry.MediaTypeResource)+"+json")

	var res specregistry.Resource
	if err := c.do(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Permanently deletes a resource.
func (c *Client) DeleteResource(ctx context.Context, namespace, resource string) error {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource)
	req, err := c.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// Lists all resources in a namespace.
func (c *Client) ListResources(ctx context.Context, namespace string) (*specregistry.ResourceList, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources")
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(specregistry.MediaTypeResourceList)+"+json")

	var list specregistry.ResourceList
	if err := c.do(req, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Creates a new version for a resource.
func (c *Client) CreateVersion(ctx context.Context, namespace, resource string, info specregistry.VersionInfo) (*specregistry.Version, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "versions")
	req, err := c.newRequest(ctx, "POST", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(specregistry.MediaTypeVersionInfo)+"+json")
	req.Header.Set("Accept", string(specregistry.MediaTypeVersion)+"+json")

	var version specregistry.Version
	if err := c.do(req, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// Retrieves version metadata with archive details.
func (c *Client) ReadVersion(ctx context.Context, namespace, resource, version string) (*specregistry.Version, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "versions", version)
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(specregistry.MediaTypeVersion)+"+json")

	var ver specregistry.Version
	if err := c.do(req, &ver); err != nil {
		return nil, err
	}
	return &ver, nil
}

// Updates mutable version metadata.
func (c *Client) UpdateVersion(ctx context.Context, namespace, resource, version string, info specregistry.VersionInfo) (*specregistry.Version, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "versions", version)
	req, err := c.newRequest(ctx, "PUT", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(specregistry.MediaTypeVersionInfo)+"+json")
	req.Header.Set("Accept", string(specregistry.MediaTypeVersion)+"+json")

	var ver specregistry.Version
	if err := c.do(req, &ver); err != nil {
		return nil, err
	}
	return &ver, nil
}

// Permanently deletes a version.
func (c *Client) DeleteVersion(ctx context.Context, namespace, resource, version string) error {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "versions", version)
	req, err := c.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// Lists all versions for a resource.
func (c *Client) ListVersions(ctx context.Context, namespace, resource string) (*specregistry.VersionList, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "versions")
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(specregistry.MediaTypeVersionList)+"+json")

	var list specregistry.VersionList
	if err := c.do(req, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Uploads a version archive.
func (c *Client) UploadArchive(ctx context.Context, namespace, resource, version string, archive io.Reader) (*specregistry.Version, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "versions", version, "archive")
	req, err := c.newRequest(ctx, "PUT", path, archive)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(specregistry.MediaTypeArchive))
	req.Header.Set("Accept", string(specregistry.MediaTypeVersion)+"+json")

	var ver specregistry.Version
	if err := c.do(req, &ver); err != nil {
		return nil, err
	}
	return &ver, nil
}

// Downloads a version archive.
func (c *Client) DownloadArchive(ctx context.Context, namespace, resource, version string) (io.ReadCloser, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "versions", version, "archive")
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(specregistry.MediaTypeArchive))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, crex.Wrap(ErrHTTPExecute, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		var regErr specregistry.Error
		if err := json.NewDecoder(resp.Body).Decode(&regErr); err != nil {
			return nil, crex.Wrapf(ErrHTTPStatus, "HTTP %d: %s", resp.StatusCode, resp.Status)
		}
		return nil, &regErr
	}

	return resp.Body, nil
}

// Creates a new channel.
func (c *Client) CreateChannel(ctx context.Context, namespace, resource string, info specregistry.ChannelInfo) (*specregistry.Channel, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "channels")
	req, err := c.newRequest(ctx, "POST", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(specregistry.MediaTypeChannelInfo)+"+json")
	req.Header.Set("Accept", string(specregistry.MediaTypeChannel)+"+json")

	var channel specregistry.Channel
	if err := c.do(req, &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

// Retrieves channel metadata with full version details.
func (c *Client) ReadChannel(ctx context.Context, namespace, resource, channel string) (*specregistry.Channel, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "channels", channel)
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(specregistry.MediaTypeChannel)+"+json")

	var ch specregistry.Channel
	if err := c.do(req, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// Updates a channel's mutable metadata.
func (c *Client) UpdateChannel(ctx context.Context, namespace, resource, channel string, info specregistry.ChannelInfo) (*specregistry.Channel, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, crex.Wrap(ErrMarshal, err)
	}

	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "channels", channel)
	req, err := c.newRequest(ctx, "PUT", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(specregistry.MediaTypeChannelInfo)+"+json")
	req.Header.Set("Accept", string(specregistry.MediaTypeChannel)+"+json")

	var ch specregistry.Channel
	if err := c.do(req, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// Permanently deletes a channel.
func (c *Client) DeleteChannel(ctx context.Context, namespace, resource, channel string) error {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "channels", channel)
	req, err := c.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// Lists all channels for a resource.
func (c *Client) ListChannels(ctx context.Context, namespace, resource string) (*specregistry.ChannelList, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "channels")
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(specregistry.MediaTypeChannelList)+"+json")

	var list specregistry.ChannelList
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
		var regErr specregistry.Error
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
