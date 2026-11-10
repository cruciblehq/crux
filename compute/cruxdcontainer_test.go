package compute

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	speccompute "github.com/cruciblehq/spec/compute"
	"github.com/cruciblehq/utils-go/codec"
)

func writeExecStream(t *testing.T, w http.ResponseWriter, msgs ...speccompute.ExecResponse) {
	t.Helper()
	w.Header().Set(speccompute.HeaderContentType, speccompute.MediaTypeExecResponse.String())
	for _, msg := range msgs {
		data, err := codec.Encode(msg, codec.JSON)
		if err != nil {
			t.Fatalf("encode: %v", err)
		}
		_, _ = w.Write(append(data, '\n'))
	}
}

func serveContainer(t *testing.T, mux *http.ServeMux) *cruxdContainer {
	t.Helper()
	return &cruxdContainer{runtime: serveCruxd(t, mux), id: "c1"}
}

func TestContainerStart(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointStart, func(w http.ResponseWriter, r *http.Request) {
		var req speccompute.StartRequest
		readJSON(t, r, &req)
		if req.ID != "c1" {
			t.Errorf("id = %q, want %q", req.ID, "c1")
		}
		w.WriteHeader(http.StatusOK)
	})
	ctr := serveContainer(t, mux)

	if err := ctr.Start(context.Background()); err != nil {
		t.Errorf("start: %v", err)
	}
}

func TestContainerStartNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointStart, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, speccompute.MediaTypeError, speccompute.Error{Code: speccompute.ErrorCodeNotFound, Message: "container not found"})
	})
	ctr := serveContainer(t, mux)

	var opErr *speccompute.Error
	if err := ctr.Start(context.Background()); !errors.As(err, &opErr) || opErr.Code != speccompute.ErrorCodeNotFound {
		t.Errorf("start = %v, want a not_found spec error", err)
	}
}

func TestContainerStop(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointStop, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ctr := serveContainer(t, mux)

	if err := ctr.Stop(context.Background()); err != nil {
		t.Errorf("stop: %v", err)
	}
}

func TestContainerCopy(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointCopy, func(w http.ResponseWriter, r *http.Request) {
		if id := r.URL.Query().Get(speccompute.QueryParamID); id != "c1" {
			t.Errorf("id = %q, want %q", id, "c1")
		}
		if ct := r.Header.Get(speccompute.HeaderContentType); ct != speccompute.MediaTypeLayer.String() {
			t.Errorf("content type = %q, want %q", ct, speccompute.MediaTypeLayer.String())
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "tar" {
			t.Errorf("body = %q, want %q", body, "tar")
		}
		w.WriteHeader(http.StatusOK)
	})
	ctr := serveContainer(t, mux)

	if err := ctr.Copy(context.Background(), bytes.NewReader([]byte("tar"))); err != nil {
		t.Errorf("copy: %v", err)
	}
}

func TestContainerConfigure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointConfig, func(w http.ResponseWriter, r *http.Request) {
		var req speccompute.ConfigureRequest
		readJSON(t, r, &req)
		if req.Config.User != "x" {
			t.Errorf("user = %q, want %q", req.Config.User, "x")
		}
		w.WriteHeader(http.StatusOK)
	})
	ctr := serveContainer(t, mux)

	if err := ctr.Configure(context.Background(), speccompute.ImageConfig{User: "x"}); err != nil {
		t.Errorf("configure: %v", err)
	}
}

func TestContainerInspect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointInspectContainer, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, speccompute.MediaTypeInspectContainerResponse, speccompute.InspectContainerResponse{Config: speccompute.ImageConfig{User: "container"}})
	})
	ctr := serveContainer(t, mux)

	cfg, err := ctr.Inspect(context.Background())
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if cfg.User != "container" {
		t.Errorf("user = %q, want %q", cfg.User, "container")
	}
}

func TestContainerCommit(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointCommit, func(w http.ResponseWriter, r *http.Request) {
		var req speccompute.CommitRequest
		readJSON(t, r, &req)
		if req.By != "me" {
			t.Errorf("by = %q, want %q", req.By, "me")
		}
		writeJSON(t, w, speccompute.MediaTypeCommitResponse, speccompute.CommitResponse{Ref: "img:committed"})
	})
	ctr := serveContainer(t, mux)

	ref, err := ctr.Commit(context.Background(), "me")
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if ref != "img:committed" {
		t.Errorf("ref = %q, want %q", ref, "img:committed")
	}
}

func TestContainerDestroy(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointDestroy, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ctr := serveContainer(t, mux)

	if err := ctr.Destroy(context.Background()); err != nil {
		t.Errorf("destroy: %v", err)
	}
}

func TestContainerExecExitCode(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointExec, func(w http.ResponseWriter, r *http.Request) {
		var req speccompute.ExecRequest
		readJSON(t, r, &req)
		if req.Command != "false" {
			t.Errorf("command = %q, want %q", req.Command, "false")
		}
		writeExecStream(t, w, speccompute.ExecResponse{Type: speccompute.ExecExit, Code: 3})
	})
	ctr := serveContainer(t, mux)

	code, err := ctr.Exec(context.Background(), io.Discard, io.Discard, "false", speccompute.Options{})
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if code != 3 {
		t.Errorf("code = %d, want 3", code)
	}
}

func TestContainerExecOutput(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointExec, func(w http.ResponseWriter, r *http.Request) {
		writeExecStream(t, w,
			speccompute.ExecResponse{Type: speccompute.ExecOutput, Channel: speccompute.ChannelStdout, Data: []byte("out\xff")},
			speccompute.ExecResponse{Type: speccompute.ExecOutput, Channel: speccompute.ChannelStderr, Data: []byte("err")},
			speccompute.ExecResponse{Type: speccompute.ExecExit},
		)
	})
	ctr := serveContainer(t, mux)

	var stdout, stderr bytes.Buffer
	code, err := ctr.Exec(context.Background(), &stdout, &stderr, "x", speccompute.Options{})
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if code != 0 {
		t.Errorf("code = %d, want 0", code)
	}
	if stdout.String() != "out\xff" {
		t.Errorf("stdout = %q, want %q", stdout.String(), "out\xff")
	}
	if stderr.String() != "err" {
		t.Errorf("stderr = %q, want %q", stderr.String(), "err")
	}
}

func TestContainerExecError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointExec, func(w http.ResponseWriter, r *http.Request) {
		writeExecStream(t, w, speccompute.ExecResponse{
			Type:  speccompute.ExecError,
			Error: &speccompute.Error{Code: speccompute.ErrorCodePreconditionFailed, Message: "container not started"},
		})
	})
	ctr := serveContainer(t, mux)

	_, err := ctr.Exec(context.Background(), io.Discard, io.Discard, "x", speccompute.Options{})
	var opErr *speccompute.Error
	if !errors.As(err, &opErr) || opErr.Code != speccompute.ErrorCodePreconditionFailed {
		t.Errorf("exec = %v, want a precondition_failed spec error", err)
	}
}

func TestContainerExecWithoutResult(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointExec, func(w http.ResponseWriter, r *http.Request) {
		writeExecStream(t, w)
	})
	ctr := serveContainer(t, mux)

	_, err := ctr.Exec(context.Background(), io.Discard, io.Discard, "x", speccompute.Options{})
	if !errors.Is(err, ErrCruxd) {
		t.Errorf("exec = %v, want ErrCruxd", err)
	}
}

func TestContainerExecUnknownChannel(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(speccompute.EndpointExec, func(w http.ResponseWriter, r *http.Request) {
		writeExecStream(t, w, speccompute.ExecResponse{Type: speccompute.ExecOutput, Channel: "<bogus>"})
	})
	ctr := serveContainer(t, mux)

	_, err := ctr.Exec(context.Background(), io.Discard, io.Discard, "x", speccompute.Options{})
	if !errors.Is(err, ErrCruxd) {
		t.Errorf("exec = %v, want ErrCruxd", err)
	}
}
