package compute

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	speccompute "github.com/cruciblehq/spec/compute"
	"github.com/cruciblehq/utils-go/codec"
)

func writeJSON(t *testing.T, w http.ResponseWriter, mt speccompute.MediaType, v any) {
	t.Helper()
	data, err := codec.Encode(v, codec.JSON)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	w.Header().Set(speccompute.HeaderContentType, mt.JSON())
	_, _ = w.Write(data)
}

func readJSON(t *testing.T, r *http.Request, v any) {
	t.Helper()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read request: %v", err)
	}
	if err := codec.Unmarshal(data, v, codec.JSON); err != nil {
		t.Fatalf("decode request: %v", err)
	}
}

func serveCruxd(t *testing.T, mux *http.ServeMux) *cruxdRuntime {
	t.Helper()
	mux.HandleFunc(speccompute.EndpointHandshake, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, speccompute.MediaTypeHandshakeResponse, speccompute.HandshakeResponse{Version: speccompute.Version})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	rt, err := Dial(context.Background(), srv.Listener.Addr().String())
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

func TestDialIncompatibleVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, speccompute.MediaTypeHandshakeResponse, speccompute.HandshakeResponse{Version: "999"})
	}))
	defer srv.Close()

	_, err := Dial(context.Background(), srv.Listener.Addr().String())
	if !errors.Is(err, ErrIncompatibleCruxd) {
		t.Errorf("dial = %v, want ErrIncompatibleCruxd", err)
	}
}

func TestDialUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	addr := srv.Listener.Addr().String()
	srv.Close()

	_, err := Dial(context.Background(), addr)
	if !errors.Is(err, ErrCruxd) {
		t.Errorf("dial = %v, want ErrCruxd", err)
	}
}

func TestRuntimeImport(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointImport, func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get(speccompute.HeaderContentType); ct != speccompute.MediaTypeImageArchive.String() {
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
		w.Header().Set(speccompute.HeaderContentType, speccompute.MediaTypeImageArchive.String())
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
		w.Header().Set(speccompute.HeaderContentType, speccompute.MediaTypeFileTree.String())
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
		writeJSON(t, w, speccompute.MediaTypeError, speccompute.Error{Code: speccompute.ErrorCodeNotFound, Message: "image not found"})
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

func TestRuntimeTransportError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointRemove, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "disk full", http.StatusInternalServerError)
	})
	rt := serveCruxd(t, mux)

	err := rt.Remove(context.Background(), "img:1")
	if !errors.Is(err, ErrCruxd) {
		t.Errorf("remove = %v, want ErrCruxd", err)
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
