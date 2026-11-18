package compute

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"

	speccompute "github.com/cruciblehq/spec/compute"
	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/httpx"
	"github.com/cruciblehq/utils-go/mediatype"
)

// Make sure the HTTP client satisfies the machine contract.
var _ Machine = (*cruxdRuntime)(nil)

// The container runtime and units of one machine, reached through cruxd.
type Machine interface {
	speccompute.Runtime

	// Returns the units kept running on the machine.
	Units() speccompute.Units
}

// Container runtime on a machine, reached through cruxd over HTTPS.
//
// Each method sends one request to cruxd and maps its response back to the
// [speccompute.Runtime] contract.
type cruxdRuntime struct {
	base   string       // Base URL of the cruxd endpoint.
	client *http.Client // Client used for every request.
}

// Connects to cruxd at addr over TLS with cfg.
//
// addr is a host and port. cfg decides which machine certificate is accepted.
// revocations is the DER revocation list sent in the handshake, so the machine
// learns about revoked certificates. Returns an error when cruxd cannot be
// reached or reports a protocol version other than [speccompute.Version].
func Dial(ctx context.Context, addr string, cfg *tls.Config, revocations []byte) (Machine, error) {
	r := &cruxdRuntime{
		base: (&url.URL{Scheme: "https", Host: addr}).String(),
		client: &http.Client{Transport: &http.Transport{
			TLSClientConfig:   cfg,
			ForceAttemptHTTP2: true,
		}},
	}
	var resp speccompute.HandshakeResponse
	req := speccompute.HandshakeRequest{Version: speccompute.Version, Revocations: revocations}
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

// Checks that cruxd and the runtime behind it can serve requests.
func (r *cruxdRuntime) Health(ctx context.Context) error {
	resp, err := r.request(ctx, http.MethodGet, speccompute.EndpointHealth, nil, "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkResponse(resp, "")
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
func (r *cruxdRuntime) call(ctx context.Context, endpoint string, reqType mediatype.MediaType, req any, respType mediatype.MediaType, out any) error {
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

// Encodes req in the compute body format and posts it to endpoint.
//
// The caller must close the response body.
func (r *cruxdRuntime) send(ctx context.Context, endpoint string, reqType mediatype.MediaType, req any) (*http.Response, error) {
	body, err := codec.Encode(req, speccompute.BodyFormat)
	if err != nil {
		return nil, crex.Wrap(ErrCruxd, err)
	}
	return r.post(ctx, endpoint, nil, reqType.With(speccompute.BodyFormat), bytes.NewReader(body))
}

// Posts body to endpoint with the given query and content type.
//
// The caller must close the response body.
func (r *cruxdRuntime) post(ctx context.Context, endpoint string, query url.Values, contentType string, body io.Reader) (*http.Response, error) {
	return r.request(ctx, http.MethodPost, endpoint, query, contentType, body)
}

// Sends a request with the given method to endpoint.
//
// The content type header is set only when contentType is not empty. The
// caller must close the response body.
func (r *cruxdRuntime) request(ctx context.Context, method, endpoint string, query url.Values, contentType string, body io.Reader) (*http.Response, error) {
	req, err := httpx.NewRequest(ctx, method, r.base, endpoint, query, body)
	if err != nil {
		return nil, crex.Wrap(ErrCruxd, err)
	}
	if contentType != "" {
		req.Header.Set(httpx.HeaderContentType, contentType)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, crex.Wrap(ErrCruxd, err)
	}
	return resp, nil
}

// Checks resp for a failure and decodes its body into out.
func decodeResponse(resp *http.Response, mt mediatype.MediaType, out any) error {
	if err := checkResponse(resp, mt.With(speccompute.BodyFormat)); err != nil {
		return err
	}
	return decodeBody(resp.Body, out)
}

// Returns the failure carried by resp, if any.
//
// A [speccompute.Error] payload is an operation failure, whatever the status,
// and is returned wrapped so callers can inspect it. Any other response with a
// status other than 200 is a transport failure. When expected is not empty,
// the response must carry that content type.
func checkResponse(resp *http.Response, expected string) error {
	ct := resp.Header.Get(httpx.HeaderContentType)
	if mediatype.Matches(ct, speccompute.MediaTypeError.With(speccompute.BodyFormat)) {
		var e speccompute.Error
		if err := decodeBody(io.LimitReader(resp.Body, httpx.MaxErrorPayload), &e); err != nil {
			return err
		}
		return crex.Wrap(ErrCruxd, &e)
	}
	if resp.StatusCode != http.StatusOK {
		return crex.Newf(ErrCruxd, "cruxd returned status %d with %q", resp.StatusCode, httpx.ErrorBody(resp))
	}
	if expected != "" && !mediatype.Matches(ct, expected) {
		return crex.Newf(ErrCruxd, "%q does not match %q", ct, expected)
	}
	return nil
}

// Reads r to the end and decodes it into out.
func decodeBody(r io.Reader, out any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return crex.Wrap(ErrCruxd, err)
	}
	if err := codec.Decode(data, out, speccompute.BodyFormat); err != nil {
		return crex.Wrap(ErrCruxd, err)
	}
	return nil
}
