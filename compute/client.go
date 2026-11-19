package compute

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"

	"github.com/cruciblehq/spec/agent"
	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/httpx"
	"github.com/cruciblehq/utils-go/mediatype"
)

// Make sure the HTTP client satisfies the agent contract.
var _ agent.Agent = (*client)(nil)

// Client for cruxd, the agent on a machine, over HTTPS.
//
// Each method sends one request to cruxd and maps its response back to the
// [agent.Agent] contract.
type client struct {
	base   string       // Base URL of the cruxd endpoint.
	client *http.Client // Client used for every request.
}

// Connects to cruxd at addr over TLS with cfg.
//
// addr is a host and port. cfg decides which agent certificate is accepted.
// revocations is the DER revocation list sent in the handshake, so the agent
// learns about revoked certificates. Returns an error when cruxd cannot be
// reached or reports a protocol version other than [agent.Version].
func Dial(ctx context.Context, addr string, cfg *tls.Config, revocations []byte) (agent.Agent, error) {
	r := &client{
		base: (&url.URL{Scheme: "https", Host: addr}).String(),
		client: &http.Client{Transport: &http.Transport{
			TLSClientConfig:   cfg,
			ForceAttemptHTTP2: true,
		}},
	}
	var resp agent.HandshakeResponse
	req := agent.HandshakeRequest{Version: agent.Version, Revocations: revocations}
	if err := r.call(ctx, agent.EndpointHandshake, agent.MediaTypeHandshakeRequest, req, agent.MediaTypeHandshakeResponse, &resp); err != nil {
		return nil, err
	}
	if resp.Version != agent.Version {
		return nil, crex.Newf(ErrIncompatibleCruxd, "cruxd speaks protocol version %q where version %q is expected", resp.Version, agent.Version)
	}
	return r, nil
}

// Imports an OCI image archive into the agent's image store.
func (r *client) Import(ctx context.Context, archive io.Reader) (agent.ImageRef, error) {
	resp, err := r.sendData(ctx, agent.EndpointImport, agent.MediaTypeImportRequest, agent.ImportRequest{}, agent.MediaTypeImageArchive, archive)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out agent.ImportResponse
	if err := decodeResponse(resp, agent.MediaTypeImportResponse, &out); err != nil {
		return "", err
	}
	return out.Ref, nil
}

// Exports the named image as an OCI tar archive written to w.
func (r *client) Export(ctx context.Context, ref agent.ImageRef, w io.Writer) error {
	resp, err := r.send(ctx, agent.EndpointExport, agent.MediaTypeExportRequest, agent.ExportRequest{Ref: ref})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	archive, err := decodeData(resp, agent.MediaTypeExportResponse, &agent.ExportResponse{}, agent.MediaTypeImageArchive)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, archive); err != nil {
		return crex.Wrap(ErrCruxd, err)
	}
	return nil
}

// Extracts a directory tree from the named image as a tar stream.
//
// The returned [io.ReadCloser] streams the data part of the response and must
// be closed by the caller.
func (r *client) Extract(ctx context.Context, ref agent.ImageRef, path string) (io.ReadCloser, error) {
	resp, err := r.send(ctx, agent.EndpointExtract, agent.MediaTypeExtractRequest, agent.ExtractRequest{Ref: ref, Path: path})
	if err != nil {
		return nil, err
	}
	tree, err := decodeData(resp, agent.MediaTypeExtractResponse, &agent.ExtractResponse{}, agent.MediaTypeFileTree)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	return struct {
		io.Reader
		io.Closer
	}{tree, resp.Body}, nil
}

// Returns the OCI image config for the named image.
func (r *client) Inspect(ctx context.Context, ref agent.ImageRef) (agent.ImageConfig, error) {
	var out agent.InspectImageResponse
	req := agent.InspectImageRequest{Ref: ref}
	if err := r.call(ctx, agent.EndpointInspect, agent.MediaTypeInspectImageRequest, req, agent.MediaTypeInspectImageResponse, &out); err != nil {
		return agent.ImageConfig{}, err
	}
	return out.Config, nil
}

// Removes the named image from the agent's image store.
func (r *client) Remove(ctx context.Context, ref agent.ImageRef) error {
	return r.call(ctx, agent.EndpointRemove, agent.MediaTypeRemoveRequest, agent.RemoveRequest{Ref: ref}, agent.MediaTypeRemoveResponse, &agent.RemoveResponse{})
}

// Creates a new container from a previously imported image.
func (r *client) Load(ctx context.Context, ref agent.ImageRef, opts agent.Options) (agent.Container, error) {
	var out agent.LoadResponse
	req := agent.LoadRequest{Ref: ref, Options: opts}
	if err := r.call(ctx, agent.EndpointLoad, agent.MediaTypeLoadRequest, req, agent.MediaTypeLoadResponse, &out); err != nil {
		return nil, err
	}
	return &container{client: r, id: out.ID}, nil
}

// Checks that cruxd and the services it depends on can serve requests.
func (r *client) Health(ctx context.Context) error {
	return r.call(ctx, agent.EndpointHealth, agent.MediaTypeHealthRequest, agent.HealthRequest{}, agent.MediaTypeHealthResponse, &agent.HealthResponse{})
}

// Releases idle connections to cruxd.
//
// Does not affect images or containers on the machine. Always returns nil.
func (r *client) Close() error {
	r.client.CloseIdleConnections()
	return nil
}

// Sends req to endpoint and decodes the response into out.
func (r *client) call(ctx context.Context, endpoint string, reqType mediatype.MediaType, req any, respType mediatype.MediaType, out any) error {
	resp, err := r.send(ctx, endpoint, reqType, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, respType, out)
}

// Encodes req in [agent.BodyFormat] and posts it to endpoint.
//
// The caller must close the response body.
func (r *client) send(ctx context.Context, endpoint string, reqType mediatype.MediaType, req any) (*http.Response, error) {
	body, err := codec.Encode(req, agent.BodyFormat)
	if err != nil {
		return nil, crex.Wrap(ErrCruxd, err)
	}
	return r.post(ctx, endpoint, reqType.With(agent.BodyFormat), bytes.NewReader(body))
}

// Posts msg as the message part and data as the data part of a multipart body.
//
// The body is streamed, so data is not held in memory. The caller must close
// the response body.
func (r *client) sendData(ctx context.Context, endpoint string, mt mediatype.MediaType, msg any, dataType mediatype.MediaType, data io.Reader) (*http.Response, error) {
	body, pw := io.Pipe()
	mw := httpx.NewMultipartWriter(pw, mt, agent.BodyFormat)
	go func() {
		pw.CloseWithError(mw.Write(msg, dataType, data))
	}()
	return r.post(ctx, endpoint, mw.ContentType(), body)
}

// Posts body to endpoint with the given content type.
//
// The caller must close the response body.
func (r *client) post(ctx context.Context, endpoint, contentType string, body io.Reader) (*http.Response, error) {
	req, err := httpx.NewRequest(ctx, http.MethodPost, r.base, endpoint, nil, body)
	if err != nil {
		return nil, crex.Wrap(ErrCruxd, err)
	}
	req.Header.Set(httpx.HeaderContentType, contentType)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, crex.Wrap(ErrCruxd, err)
	}
	return resp, nil
}

// Checks resp for a failure and decodes its body into out.
func decodeResponse(resp *http.Response, mt mediatype.MediaType, out any) error {
	if err := checkResponse(resp, mt.With(agent.BodyFormat)); err != nil {
		return err
	}
	return decodeBody(resp.Body, out)
}

// Checks resp for a failure, decodes its message part into out, and returns
// its data part.
//
// The data part reads from the response body, which must stay open until the
// data part is read.
func decodeData(resp *http.Response, mt mediatype.MediaType, out any, dataType mediatype.MediaType) (io.Reader, error) {
	if err := checkResponse(resp, ""); err != nil {
		return nil, err
	}
	data, err := httpx.ReadMultipart(resp.Header.Get(httpx.HeaderContentType), resp.Body, mt, agent.BodyFormat, out, dataType)
	if err != nil {
		return nil, crex.Wrap(ErrCruxd, err)
	}
	return data, nil
}

// Returns the failure carried by resp, if any.
//
// An [agent.Error] payload is an operation failure, whatever the status, and
// is returned wrapped so callers can inspect it. Any other response with a
// status other than 200 is a transport failure. When expected is not empty,
// the response must carry that content type.
func checkResponse(resp *http.Response, expected string) error {
	ct := resp.Header.Get(httpx.HeaderContentType)
	if mediatype.Matches(ct, agent.MediaTypeError.With(agent.BodyFormat)) {
		var e agent.Error
		if err := decodeBody(io.LimitReader(resp.Body, httpx.MaxErrorPayload), &e); err != nil {
			return err
		}
		return crex.Wrap(ErrCruxd, &e)
	}
	if resp.StatusCode != http.StatusOK {
		return crex.Newf(ErrCruxd, "cruxd returned status %d with %q", resp.StatusCode, httpx.ErrorBody(resp))
	}
	if expected != "" && !mediatype.Matches(ct, expected) {
		return crex.Newf(ErrCruxd, "cruxd responded with media type %q where %q is expected", ct, expected)
	}
	return nil
}

// Reads r to the end and decodes it into out.
func decodeBody(r io.Reader, out any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return crex.Wrap(ErrCruxd, err)
	}
	if err := codec.Decode(data, out, agent.BodyFormat); err != nil {
		return crex.Wrap(ErrCruxd, err)
	}
	return nil
}
