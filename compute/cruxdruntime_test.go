package compute

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	speccompute "github.com/cruciblehq/spec/compute"
	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/httpx"
	"github.com/cruciblehq/utils-go/mediatype"
)

func writeJSON(t *testing.T, w http.ResponseWriter, mt mediatype.MediaType, v any) {
	t.Helper()
	if err := httpx.Write(w, mt, speccompute.BodyFormat, http.StatusOK, v); err != nil {
		t.Fatalf("write response: %v", err)
	}
}

func writeError(t *testing.T, w http.ResponseWriter, status int, code speccompute.ErrorCode, message string) {
	t.Helper()
	if err := httpx.Write(w, speccompute.MediaTypeError, speccompute.BodyFormat, status, speccompute.Error{Code: code, Message: message}); err != nil {
		t.Fatalf("write error: %v", err)
	}
}

func readJSON(t *testing.T, r *http.Request, v any) {
	t.Helper()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read request: %v", err)
	}
	if err := codec.Decode(data, v, speccompute.BodyFormat); err != nil {
		t.Fatalf("decode request: %v", err)
	}
}

func tlsConfig(srv *httptest.Server) *tls.Config {
	return srv.Client().Transport.(*http.Transport).TLSClientConfig
}

func serveCruxd(t *testing.T, mux *http.ServeMux) *cruxdRuntime {
	t.Helper()
	mux.HandleFunc(speccompute.EndpointHandshake, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, speccompute.MediaTypeHandshakeResponse, speccompute.HandshakeResponse{Version: speccompute.Version})
	})
	srv := httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)
	rt, err := Dial(context.Background(), srv.Listener.Addr().String(), tlsConfig(srv), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return rt.(*cruxdRuntime)
}

func TestDial(t *testing.T) {
	rt := serveCruxd(t, http.NewServeMux())
	if err := rt.Close(); err != nil {
		t.Errorf("close: %v", err)
	}
}

func TestDialSendsRevocations(t *testing.T) {
	var got speccompute.HandshakeRequest
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readJSON(t, r, &got)
		writeJSON(t, w, speccompute.MediaTypeHandshakeResponse, speccompute.HandshakeResponse{Version: speccompute.Version})
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
		writeJSON(t, w, speccompute.MediaTypeHandshakeResponse, speccompute.HandshakeResponse{Version: "999"})
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
	mux.HandleFunc(speccompute.EndpointImport, func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get(httpx.HeaderContentType); ct != speccompute.MediaTypeImageArchive.String() {
			t.Errorf("content type = %q, want %q", ct, speccompute.MediaTypeImageArchive.String())
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "archive" {
			t.Errorf("body = %q, want %q", body, "archive")
		}
		writeJSON(t, w, speccompute.MediaTypeImportResponse, speccompute.ImportResponse{Ref: "img:1"})
	})
	rt := serveCruxd(t, mux)

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
	mux.HandleFunc(speccompute.EndpointExport, func(w http.ResponseWriter, r *http.Request) {
		var req speccompute.ExportRequest
		readJSON(t, r, &req)
		if req.Ref != "img:1" {
			t.Errorf("ref = %q, want %q", req.Ref, "img:1")
		}
		w.Header().Set(httpx.HeaderContentType, speccompute.MediaTypeImageArchive.String())
		_, _ = w.Write([]byte("archive"))
	})
	rt := serveCruxd(t, mux)

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
	mux.HandleFunc(speccompute.EndpointExtract, func(w http.ResponseWriter, r *http.Request) {
		var req speccompute.ExtractRequest
		readJSON(t, r, &req)
		if req.Path != "/app" {
			t.Errorf("path = %q, want %q", req.Path, "/app")
		}
		w.Header().Set(httpx.HeaderContentType, speccompute.MediaTypeFileTree.String())
		_, _ = w.Write([]byte("tar"))
	})
	rt := serveCruxd(t, mux)

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
	mux.HandleFunc(speccompute.EndpointInspect, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, speccompute.MediaTypeInspectImageResponse, speccompute.InspectImageResponse{Config: speccompute.ImageConfig{User: "image"}})
	})
	rt := serveCruxd(t, mux)

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
	mux.HandleFunc(speccompute.EndpointRemove, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	rt := serveCruxd(t, mux)

	if err := rt.Remove(context.Background(), "img:1"); err != nil {
		t.Errorf("remove: %v", err)
	}
}

func TestRuntimeLoad(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointLoad, func(w http.ResponseWriter, r *http.Request) {
		var req speccompute.LoadRequest
		readJSON(t, r, &req)
		if !req.Options.HostNetwork {
			t.Error("host network option was not sent")
		}
		writeJSON(t, w, speccompute.MediaTypeLoadResponse, speccompute.LoadResponse{ID: "c1"})
	})
	rt := serveCruxd(t, mux)

	ctr, err := rt.Load(context.Background(), "img:1", speccompute.Options{HostNetwork: true})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if id := ctr.(*cruxdContainer).id; id != "c1" {
		t.Errorf("id = %q, want %q", id, "c1")
	}
}

func TestRuntimeOperationError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointInspect, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusNotFound, speccompute.ErrorCodeNotFound, "image not found")
	})
	rt := serveCruxd(t, mux)

	_, err := rt.Inspect(context.Background(), "img:missing")
	if !errors.Is(err, ErrCruxd) {
		t.Fatalf("inspect = %v, want ErrCruxd", err)
	}
	var opErr *speccompute.Error
	if !errors.As(err, &opErr) {
		t.Fatalf("inspect = %v, want a spec error", err)
	}
	if opErr.Code != speccompute.ErrorCodeNotFound {
		t.Errorf("code = %q, want %q", opErr.Code, speccompute.ErrorCodeNotFound)
	}
}

func TestRuntimeOperationErrorAtStatusOK(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointInspect, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusOK, speccompute.ErrorCodeNotFound, "image not found")
	})
	rt := serveCruxd(t, mux)

	_, err := rt.Inspect(context.Background(), "img:missing")
	var opErr *speccompute.Error
	if !errors.As(err, &opErr) || opErr.Code != speccompute.ErrorCodeNotFound {
		t.Fatalf("inspect = %v, want a not_found spec error", err)
	}
}

func TestRuntimeTransportError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointRemove, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "disk full", http.StatusInternalServerError)
	})
	rt := serveCruxd(t, mux)

	err := rt.Remove(context.Background(), "img:1")
	var opErr *speccompute.Error
	if !errors.Is(err, ErrCruxd) || errors.As(err, &opErr) {
		t.Errorf("remove = %v, want a transport ErrCruxd", err)
	}
}

func TestRuntimeHealth(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointHealth, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want %q", r.Method, http.MethodGet)
		}
		w.WriteHeader(http.StatusOK)
	})
	rt := serveCruxd(t, mux)

	if err := rt.Health(context.Background()); err != nil {
		t.Errorf("health: %v", err)
	}
}

func TestRuntimeHealthUnavailable(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointHealth, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "containerd is not serving", http.StatusServiceUnavailable)
	})
	rt := serveCruxd(t, mux)

	if err := rt.Health(context.Background()); !errors.Is(err, ErrCruxd) {
		t.Errorf("health = %v, want ErrCruxd", err)
	}
}

func TestRuntimeUnexpectedContentType(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointLoad, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, speccompute.MediaTypeCommitResponse, speccompute.CommitResponse{Ref: "x"})
	})
	rt := serveCruxd(t, mux)

	_, err := rt.Load(context.Background(), "img:1", speccompute.Options{})
	if !errors.Is(err, ErrCruxd) {
		t.Errorf("load = %v, want ErrCruxd", err)
	}
}
