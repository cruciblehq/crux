package compute

import (
	"bytes"
	"context"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	speccompute "github.com/cruciblehq/spec/compute"
	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/crex"
)

// Largest transport error body read into an error message.
const maxErrorBody = 4096

// Make sure the HTTP client satisfies the spec.
var _ speccompute.Runtime = (*cruxdRuntime)(nil)

// Container runtime on a machine, reached through cruxd over HTTP.
//
// Each method sends one request to cruxd and maps its response back to the
// [speccompute.Runtime] contract.
type cruxdRuntime struct {
	base   string       // Base URL of the cruxd endpoint.
	client *http.Client // Client used for every request.
}

// Connects to cruxd at addr.
//
// addr is a host and port. Returns an error when cruxd cannot be reached or
// reports a protocol version other than [speccompute.Version].
func Dial(ctx context.Context, addr string) (speccompute.Runtime, error) {
	r := &cruxdRuntime{
		base:   (&url.URL{Scheme: "http", Host: addr}).String(),
		client: &http.Client{},
	}
	var resp speccompute.HandshakeResponse
	req := speccompute.HandshakeRequest{Version: speccompute.Version}
	if err := r.call(ctx, speccompute.EndpointHandshake, speccompute.MediaTypeHandshakeRequest, req, speccompute.MediaTypeHandshakeResponse, &resp); err != nil {
		return nil, err
	}
	if resp.Version != speccompute.Version {
		return nil, crex.Newf(ErrIncompatibleCruxd, "%q does not match %q", resp.Version, speccompute.Version)
	}
	return r, nil
}

// Imports an OCI image archive into the machine's image store.
func (r *cruxdRuntime) Import(ctx context.Context, archive io.Reader) (speccompute.ImageRef, error) {
	resp, err := r.post(ctx, speccompute.EndpointImport, nil, speccompute.MediaTypeImageArchive.String(), archive)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out speccompute.ImportResponse
	if err := decodeResponse(resp, speccompute.MediaTypeImportResponse, &out); err != nil {
		return "", err
	}
	return out.Ref, nil
}

// Exports the named image as an OCI tar archive written to w.
func (r *cruxdRuntime) Export(ctx context.Context, ref speccompute.ImageRef, w io.Writer) error {
	resp, err := r.send(ctx, speccompute.EndpointExport, speccompute.MediaTypeExportRequest, speccompute.ExportRequest{Ref: ref})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := checkResponse(resp, speccompute.MediaTypeImageArchive.String()); err != nil {
		return err
	}
	if _, err := io.Copy(w, resp.Body); err != nil {
		return crex.Wrap(ErrCruxd, err)
	}
	return nil
}

// Extracts a directory tree from the named image as a tar stream.
//
// The returned [io.ReadCloser] streams the response body and must be closed
// by the caller.
func (r *cruxdRuntime) Extract(ctx context.Context, ref speccompute.ImageRef, path string) (io.ReadCloser, error) {
	resp, err := r.send(ctx, speccompute.EndpointExtract, speccompute.MediaTypeExtractRequest, speccompute.ExtractRequest{Ref: ref, Path: path})
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp, speccompute.MediaTypeFileTree.String()); err != nil {
		resp.Body.Close()
		return nil, err
	}
	return resp.Body, nil
}

// Returns the OCI image config for the named image.
func (r *cruxdRuntime) Inspect(ctx context.Context, ref speccompute.ImageRef) (speccompute.ImageConfig, error) {
	var out speccompute.InspectImageResponse
	req := speccompute.InspectImageRequest{Ref: ref}
	if err := r.call(ctx, speccompute.EndpointInspect, speccompute.MediaTypeInspectImageRequest, req, speccompute.MediaTypeInspectImageResponse, &out); err != nil {
		return speccompute.ImageConfig{}, err
	}
	return out.Config, nil
}

// Removes the named image from the store.
func (r *cruxdRuntime) Remove(ctx context.Context, ref speccompute.ImageRef) error {
	return r.call(ctx, speccompute.EndpointRemove, speccompute.MediaTypeRemoveRequest, speccompute.RemoveRequest{Ref: ref}, "", nil)
}

// Creates a new container from a previously imported image.
func (r *cruxdRuntime) Load(ctx context.Context, ref speccompute.ImageRef, opts speccompute.Options) (speccompute.Container, error) {
	var out speccompute.LoadResponse
	req := speccompute.LoadRequest{Ref: ref, Options: opts}
	if err := r.call(ctx, speccompute.EndpointLoad, speccompute.MediaTypeLoadRequest, req, speccompute.MediaTypeLoadResponse, &out); err != nil {
		return nil, err
	}
	return &cruxdContainer{runtime: r, id: out.ID}, nil
}

// Releases idle connections to cruxd.
//
// Does not affect images or containers on the machine. Always returns nil.
func (r *cruxdRuntime) Close() error {
	r.client.CloseIdleConnections()
	return nil
}

// Sends req to endpoint and decodes the response into out.
//
// When out is nil the response is expected to have no body.
func (r *cruxdRuntime) call(ctx context.Context, endpoint string, reqType speccompute.MediaType, req any, respType speccompute.MediaType, out any) error {
	resp, err := r.send(ctx, endpoint, reqType, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out == nil {
		return checkResponse(resp, "")
	}
	return decodeResponse(resp, respType, out)
}

// Encodes req as JSON and posts it to endpoint.
//
// The caller must close the response body.
func (r *cruxdRuntime) send(ctx context.Context, endpoint string, reqType speccompute.MediaType, req any) (*http.Response, error) {
	body, err := codec.Encode(req, codec.JSON)
	if err != nil {
		return nil, crex.Wrap(ErrCruxd, err)
	}
	return r.post(ctx, endpoint, nil, reqType.JSON(), bytes.NewReader(body))
}

// Posts body to endpoint with the given query and content type.
//
// The caller must close the response body.
func (r *cruxdRuntime) post(ctx context.Context, endpoint string, query url.Values, contentType string, body io.Reader) (*http.Response, error) {
	u, err := url.Parse(r.base)
	if err != nil {
		return nil, crex.Wrap(ErrCruxd, err)
	}
	u.Path = endpoint
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), body)
	if err != nil {
		return nil, crex.Wrap(ErrCruxd, err)
	}
	req.Header.Set(speccompute.HeaderContentType, contentType)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, crex.Wrap(ErrCruxd, err)
	}
	return resp, nil
}

// Checks resp for a failure and decodes its JSON body into out.
func decodeResponse(resp *http.Response, mt speccompute.MediaType, out any) error {
	if err := checkResponse(resp, mt.JSON()); err != nil {
		return err
	}
	return decodeBody(resp.Body, out)
}

// Returns the failure carried by resp, if any.
//
// A non-200 status is a transport failure. A [speccompute.Error] payload is an
// operation failure and is returned wrapped so callers can inspect it. When
// expected is not empty, the response must carry that content type.
func checkResponse(resp *http.Response, expected string) error {
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return crex.Newf(ErrCruxd, "cruxd returned status %d with %q", resp.StatusCode, strings.TrimSpace(string(msg)))
	}
	ct := resp.Header.Get(speccompute.HeaderContentType)
	mt, _, _ := mime.ParseMediaType(ct)
	if mt == speccompute.MediaTypeError.JSON() {
		var e speccompute.Error
		if err := decodeBody(resp.Body, &e); err != nil {
			return err
		}
		return crex.Wrap(ErrCruxd, &e)
	}
	if expected != "" && mt != expected {
		return crex.Newf(ErrCruxd, "%q does not match %q", ct, expected)
	}
	return nil
}

// Reads r to the end and decodes it as JSON into out.
func decodeBody(r io.Reader, out any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return crex.Wrap(ErrCruxd, err)
	}
	if err := codec.Unmarshal(data, out, codec.JSON); err != nil {
		return crex.Wrap(ErrCruxd, err)
	}
	return nil
}
