package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
func (c *Client) CreateNamespace(ctx context.Context, info NamespaceInfo) (*Namespace, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, fmt.Errorf("marshal namespace info: %w", err)
	}

	req, err := c.newRequest(ctx, "POST", "/namespaces", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(MediaTypeNamespaceInfo)+"+json")
	req.Header.Set("Accept", string(MediaTypeNamespace)+"+json")

	var ns Namespace
	if err := c.do(req, &ns); err != nil {
		return nil, err
	}
	return &ns, nil
}

// Retrieves namespace metadata and resource summaries.
func (c *Client) ReadNamespace(ctx context.Context, namespace string) (*Namespace, error) {
	path, _ := url.JoinPath("/namespaces", namespace)
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(MediaTypeNamespace)+"+json")

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
		return nil, fmt.Errorf("marshal namespace info: %w", err)
	}

	path, _ := url.JoinPath("/namespaces", namespace)
	req, err := c.newRequest(ctx, "PUT", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(MediaTypeNamespaceInfo)+"+json")
	req.Header.Set("Accept", string(MediaTypeNamespace)+"+json")

	var ns Namespace
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
func (c *Client) ListNamespaces(ctx context.Context) (*NamespaceList, error) {
	req, err := c.newRequest(ctx, "GET", "/namespaces", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(MediaTypeNamespaceList)+"+json")

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
		return nil, fmt.Errorf("marshal resource info: %w", err)
	}

	path, _ := url.JoinPath("/namespaces", namespace, "resources")
	req, err := c.newRequest(ctx, "POST", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(MediaTypeResourceInfo)+"+json")
	req.Header.Set("Accept", string(MediaTypeResource)+"+json")

	var resource Resource
	if err := c.do(req, &resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

// Retrieves resource metadata with version and channel summaries.
func (c *Client) ReadResource(ctx context.Context, namespace, resource string) (*Resource, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource)
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(MediaTypeResource)+"+json")

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
		return nil, fmt.Errorf("marshal resource info: %w", err)
	}

	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource)
	req, err := c.newRequest(ctx, "PUT", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(MediaTypeResourceInfo)+"+json")
	req.Header.Set("Accept", string(MediaTypeResource)+"+json")

	var res Resource
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
func (c *Client) ListResources(ctx context.Context, namespace string) (*ResourceList, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources")
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(MediaTypeResourceList)+"+json")

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
		return nil, fmt.Errorf("marshal version info: %w", err)
	}

	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "versions")
	req, err := c.newRequest(ctx, "POST", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(MediaTypeVersionInfo)+"+json")
	req.Header.Set("Accept", string(MediaTypeVersion)+"+json")

	var version Version
	if err := c.do(req, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// Retrieves version metadata with archive details.
func (c *Client) ReadVersion(ctx context.Context, namespace, resource, version string) (*Version, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "versions", version)
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(MediaTypeVersion)+"+json")

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
		return nil, fmt.Errorf("marshal version info: %w", err)
	}

	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "versions", version)
	req, err := c.newRequest(ctx, "PUT", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(MediaTypeVersionInfo)+"+json")
	req.Header.Set("Accept", string(MediaTypeVersion)+"+json")

	var ver Version
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
func (c *Client) ListVersions(ctx context.Context, namespace, resource string) (*VersionList, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "versions")
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(MediaTypeVersionList)+"+json")

	var list VersionList
	if err := c.do(req, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Uploads a version archive.
func (c *Client) UploadArchive(ctx context.Context, namespace, resource, version string, archive io.Reader) (*Version, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "versions", version, "archive")
	req, err := c.newRequest(ctx, "PUT", path, archive)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(MediaTypeArchive))
	req.Header.Set("Accept", string(MediaTypeVersion)+"+json")

	var ver Version
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
	req.Header.Set("Accept", string(MediaTypeArchive))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		var regErr Error
		if err := json.NewDecoder(resp.Body).Decode(&regErr); err != nil {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}
		return nil, &regErr
	}

	return resp.Body, nil
}

// Creates a new channel.
func (c *Client) CreateChannel(ctx context.Context, namespace, resource string, info ChannelInfo) (*Channel, error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, fmt.Errorf("marshal channel info: %w", err)
	}

	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "channels")
	req, err := c.newRequest(ctx, "POST", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(MediaTypeChannelInfo)+"+json")
	req.Header.Set("Accept", string(MediaTypeChannel)+"+json")

	var channel Channel
	if err := c.do(req, &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

// Retrieves channel metadata with full version details.
func (c *Client) ReadChannel(ctx context.Context, namespace, resource, channel string) (*Channel, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "channels", channel)
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(MediaTypeChannel)+"+json")

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
		return nil, fmt.Errorf("marshal channel info: %w", err)
	}

	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "channels", channel)
	req, err := c.newRequest(ctx, "PUT", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", string(MediaTypeChannelInfo)+"+json")
	req.Header.Set("Accept", string(MediaTypeChannel)+"+json")

	var ch Channel
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
func (c *Client) ListChannels(ctx context.Context, namespace, resource string) (*ChannelList, error) {
	path, _ := url.JoinPath("/namespaces", namespace, "resources", resource, "channels")
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", string(MediaTypeChannelList)+"+json")

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
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	u.Path = path

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	return req, nil
}

// Executes an HTTP request and decodes the JSON response.
func (c *Client) do(req *http.Request, result interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var regErr Error
		if err := json.NewDecoder(resp.Body).Decode(&regErr); err != nil {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}
		return &regErr
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
