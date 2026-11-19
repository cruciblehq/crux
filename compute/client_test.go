package compute

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cruciblehq/spec/agent"
	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/httpx"
	"github.com/cruciblehq/utils-go/mediatype"
)

func writeJSON(t *testing.T, w http.ResponseWriter, mt mediatype.MediaType, v any) {
	t.Helper()
	if err := httpx.Write(w, mt, agent.BodyFormat, http.StatusOK, v); err != nil {
		t.Fatalf("write response: %v", err)
	}
}

func writeError(t *testing.T, w http.ResponseWriter, status int, code agent.ErrorCode, message string) {
	t.Helper()
	if err := httpx.Write(w, agent.MediaTypeError, agent.BodyFormat, status, agent.Error{Code: code, Message: message}); err != nil {
		t.Fatalf("write error: %v", err)
	}
}

func readJSON(t *testing.T, r *http.Request, v any) {
	t.Helper()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read request: %v", err)
	}
	if err := codec.Decode(data, v, agent.BodyFormat); err != nil {
		t.Fatalf("decode request: %v", err)
	}
}

func readData(t *testing.T, r *http.Request, mt mediatype.MediaType, v any, dataType mediatype.MediaType) string {
	t.Helper()
	data, err := httpx.ReadMultipart(r.Header.Get(httpx.HeaderContentType), r.Body, mt, agent.BodyFormat, v, dataType)
	if err != nil {
		t.Fatalf("read multipart request: %v", err)
	}
	b, err := io.ReadAll(data)
	if err != nil {
		t.Fatalf("read data: %v", err)
	}
	return string(b)
}

func writeData(t *testing.T, w http.ResponseWriter, mt mediatype.MediaType, v any, dataType mediatype.MediaType, data string) {
	t.Helper()
	mw := httpx.NewMultipartWriter(w, mt, agent.BodyFormat)
	w.Header().Set(httpx.HeaderContentType, mw.ContentType())
	if err := mw.Write(v, dataType, strings.NewReader(data)); err != nil {
		t.Fatalf("write multipart response: %v", err)
	}
}

func tlsConfig(srv *httptest.Server) *tls.Config {
	return srv.Client().Transport.(*http.Transport).TLSClientConfig
}

func serveClient(t *testing.T, mux *http.ServeMux) *client {
	t.Helper()
	mux.HandleFunc(agent.EndpointHandshake, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, agent.MediaTypeHandshakeResponse, agent.HandshakeResponse{Version: agent.Version})
	})
	srv := httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)
	rt, err := Dial(context.Background(), srv.Listener.Addr().String(), tlsConfig(srv), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return rt.(*client)
}

func TestDial(t *testing.T) {
	rt := serveClient(t, http.NewServeMux())
	if err := rt.Close(); err != nil {
		t.Errorf("close: %v", err)
	}
}

func TestDialSendsRevocations(t *testing.T) {
	var got agent.HandshakeRequest
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readJSON(t, r, &got)
		writeJSON(t, w, agent.MediaTypeHandshakeResponse, agent.HandshakeResponse{Version: agent.Version})
	}))
	defer srv.Close()

	if _, err := Dial(context.Background(), srv.Listener.Addr().String(), tlsConfig(srv), []byte("crl")); err != nil {
		t.Fatalf("dial: %v", err)
	}
	if string(got.Revocations) != "crl" {
		t.Errorf("Revocations = %q, want %q", got.Revocations, "crl")
	}
}

func TestDialIncompatibleVersion(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, agent.MediaTypeHandshakeResponse, agent.HandshakeResponse{Version: "999"})
	}))
	defer srv.Close()

	_, err := Dial(context.Background(), srv.Listener.Addr().String(), tlsConfig(srv), nil)
	if !errors.Is(err, ErrIncompatibleCruxd) {
		t.Errorf("dial = %v, want ErrIncompatibleCruxd", err)
	}
}

func TestDialUnreachable(t *testing.T) {
	srv := httptest.NewTLSServer(http.NotFoundHandler())
	addr := srv.Listener.Addr().String()
	srv.Close()

	_, err := Dial(context.Background(), addr, tlsConfig(srv), nil)
	if !errors.Is(err, ErrCruxd) {
		t.Errorf("dial = %v, want ErrCruxd", err)
	}
}

func TestRuntimeImport(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointImport, func(w http.ResponseWriter, r *http.Request) {
		var req agent.ImportRequest
		if data := readData(t, r, agent.MediaTypeImportRequest, &req, agent.MediaTypeImageArchive); data != "archive" {
			t.Errorf("data = %q, want %q", data, "archive")
		}
		writeJSON(t, w, agent.MediaTypeImportResponse, agent.ImportResponse{Ref: "img:1"})
	})
	rt := serveClient(t, mux)

	ref, err := rt.Import(context.Background(), bytes.NewReader([]byte("archive")))
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if ref != "img:1" {
		t.Errorf("ref = %q, want %q", ref, "img:1")
	}
}

func TestRuntimeExport(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointExport, func(w http.ResponseWriter, r *http.Request) {
		var req agent.ExportRequest
		readJSON(t, r, &req)
		if req.Ref != "img:1" {
			t.Errorf("ref = %q, want %q", req.Ref, "img:1")
		}
		writeData(t, w, agent.MediaTypeExportResponse, agent.ExportResponse{}, agent.MediaTypeImageArchive, "archive")
	})
	rt := serveClient(t, mux)

	var buf bytes.Buffer
	if err := rt.Export(context.Background(), "img:1", &buf); err != nil {
		t.Fatalf("export: %v", err)
	}
	if buf.String() != "archive" {
		t.Errorf("archive = %q, want %q", buf.String(), "archive")
	}
}

func TestRuntimeExtract(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointExtract, func(w http.ResponseWriter, r *http.Request) {
		var req agent.ExtractRequest
		readJSON(t, r, &req)
		if req.Path != "/app" {
			t.Errorf("path = %q, want %q", req.Path, "/app")
		}
		writeData(t, w, agent.MediaTypeExtractResponse, agent.ExtractResponse{}, agent.MediaTypeFileTree, "tar")
	})
	rt := serveClient(t, mux)

	rc, err := rt.Extract(context.Background(), "img:1", "/app")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	defer rc.Close()
	data, _ := io.ReadAll(rc)
	if string(data) != "tar" {
		t.Errorf("tree = %q, want %q", data, "tar")
	}
}

func TestRuntimeInspect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointInspect, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, agent.MediaTypeInspectImageResponse, agent.InspectImageResponse{Config: agent.ImageConfig{User: "image"}})
	})
	rt := serveClient(t, mux)

	cfg, err := rt.Inspect(context.Background(), "img:1")
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if cfg.User != "image" {
		t.Errorf("user = %q, want %q", cfg.User, "image")
	}
}

func TestRuntimeRemove(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointRemove, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, agent.MediaTypeRemoveResponse, agent.RemoveResponse{})
	})
	rt := serveClient(t, mux)

	if err := rt.Remove(context.Background(), "img:1"); err != nil {
		t.Errorf("remove: %v", err)
	}
}

func TestRuntimeLoad(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointLoad, func(w http.ResponseWriter, r *http.Request) {
		var req agent.LoadRequest
		readJSON(t, r, &req)
		if !req.Options.HostNetwork {
			t.Error("host network option was not sent")
		}
		writeJSON(t, w, agent.MediaTypeLoadResponse, agent.LoadResponse{ID: "c1"})
	})
	rt := serveClient(t, mux)

	ctr, err := rt.Load(context.Background(), "img:1", agent.Options{HostNetwork: true})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if id := ctr.(*container).id; id != "c1" {
		t.Errorf("id = %q, want %q", id, "c1")
	}
}

func TestRuntimeOperationError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointInspect, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusNotFound, agent.ErrorCodeNotFound, "image not found")
	})
	rt := serveClient(t, mux)

	_, err := rt.Inspect(context.Background(), "img:missing")
	if !errors.Is(err, ErrCruxd) {
		t.Fatalf("inspect = %v, want ErrCruxd", err)
	}
	var opErr *agent.Error
	if !errors.As(err, &opErr) {
		t.Fatalf("inspect = %v, want a spec error", err)
	}
	if opErr.Code != agent.ErrorCodeNotFound {
		t.Errorf("code = %q, want %q", opErr.Code, agent.ErrorCodeNotFound)
	}
}

func TestRuntimeOperationErrorAtStatusOK(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointInspect, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusOK, agent.ErrorCodeNotFound, "image not found")
	})
	rt := serveClient(t, mux)

	_, err := rt.Inspect(context.Background(), "img:missing")
	var opErr *agent.Error
	if !errors.As(err, &opErr) || opErr.Code != agent.ErrorCodeNotFound {
		t.Fatalf("inspect = %v, want a not_found spec error", err)
	}
}

func TestRuntimeTransportError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointRemove, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "disk full", http.StatusInternalServerError)
	})
	rt := serveClient(t, mux)

	err := rt.Remove(context.Background(), "img:1")
	var opErr *agent.Error
	if !errors.Is(err, ErrCruxd) || errors.As(err, &opErr) {
		t.Errorf("remove = %v, want a transport ErrCruxd", err)
	}
}

func TestRuntimeHealth(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointHealth, func(w http.ResponseWriter, r *http.Request) {
		var req agent.HealthRequest
		readJSON(t, r, &req)
		writeJSON(t, w, agent.MediaTypeHealthResponse, agent.HealthResponse{})
	})
	rt := serveClient(t, mux)

	if err := rt.Health(context.Background()); err != nil {
		t.Errorf("health: %v", err)
	}
}

func TestRuntimeHealthUnavailable(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointHealth, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "containerd is not serving", http.StatusServiceUnavailable)
	})
	rt := serveClient(t, mux)

	if err := rt.Health(context.Background()); !errors.Is(err, ErrCruxd) {
		t.Errorf("health = %v, want ErrCruxd", err)
	}
}

func TestRuntimeUnexpectedContentType(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointLoad, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, agent.MediaTypeCommitResponse, agent.CommitResponse{Ref: "x"})
	})
	rt := serveClient(t, mux)

	_, err := rt.Load(context.Background(), "img:1", agent.Options{})
	if !errors.Is(err, ErrCruxd) {
		t.Errorf("load = %v, want ErrCruxd", err)
	}
}
