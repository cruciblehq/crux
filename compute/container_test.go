package compute

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/cruciblehq/spec/agent"
	"github.com/cruciblehq/utils-go/httpx"
)

func writeExecStream(t *testing.T, w http.ResponseWriter, msgs ...agent.ExecResponse) {
	t.Helper()
	w.Header().Set(httpx.HeaderContentType, agent.MediaTypeExecResponse.String())
	stream := agent.NewExecWriter(w)
	for _, msg := range msgs {
		if err := stream.Send(msg); err != nil {
			t.Fatalf("send: %v", err)
		}
	}
}

func serveContainer(t *testing.T, mux *http.ServeMux) *container {
	t.Helper()
	return &container{client: serveClient(t, mux), id: "c1"}
}

func TestContainerStart(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointStart, func(w http.ResponseWriter, r *http.Request) {
		var req agent.StartRequest
		readJSON(t, r, &req)
		if req.ID != "c1" {
			t.Errorf("id = %q, want %q", req.ID, "c1")
		}
		writeJSON(t, w, agent.MediaTypeStartResponse, agent.StartResponse{})
	})
	ctr := serveContainer(t, mux)

	if err := ctr.Start(context.Background()); err != nil {
		t.Errorf("start: %v", err)
	}
}

func TestContainerStartNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointStart, func(w http.ResponseWriter, r *http.Request) {
		writeError(t, w, http.StatusNotFound, agent.ErrorCodeNotFound, "container not found")
	})
	ctr := serveContainer(t, mux)

	var opErr *agent.Error
	if err := ctr.Start(context.Background()); !errors.As(err, &opErr) || opErr.Code != agent.ErrorCodeNotFound {
		t.Errorf("start = %v, want a not_found spec error", err)
	}
}

func TestContainerStop(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointStop, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, agent.MediaTypeStopResponse, agent.StopResponse{})
	})
	ctr := serveContainer(t, mux)

	if err := ctr.Stop(context.Background()); err != nil {
		t.Errorf("stop: %v", err)
	}
}

func TestContainerCopy(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointCopy, func(w http.ResponseWriter, r *http.Request) {
		var req agent.CopyRequest
		if data := readData(t, r, agent.MediaTypeCopyRequest, &req, agent.MediaTypeLayer); data != "tar" {
			t.Errorf("data = %q, want %q", data, "tar")
		}
		if req.ID != "c1" {
			t.Errorf("id = %q, want %q", req.ID, "c1")
		}
		writeJSON(t, w, agent.MediaTypeCopyResponse, agent.CopyResponse{})
	})
	ctr := serveContainer(t, mux)

	if err := ctr.Copy(context.Background(), bytes.NewReader([]byte("tar"))); err != nil {
		t.Errorf("copy: %v", err)
	}
}

func TestContainerConfigure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointConfig, func(w http.ResponseWriter, r *http.Request) {
		var req agent.ConfigureRequest
		readJSON(t, r, &req)
		if req.Config.User != "x" {
			t.Errorf("user = %q, want %q", req.Config.User, "x")
		}
		writeJSON(t, w, agent.MediaTypeConfigureResponse, agent.ConfigureResponse{})
	})
	ctr := serveContainer(t, mux)

	if err := ctr.Configure(context.Background(), agent.ImageConfig{User: "x"}); err != nil {
		t.Errorf("configure: %v", err)
	}
}

func TestContainerInspect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointInspectContainer, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, agent.MediaTypeInspectContainerResponse, agent.InspectContainerResponse{Config: agent.ImageConfig{User: "container"}})
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
	mux.HandleFunc(agent.EndpointCommit, func(w http.ResponseWriter, r *http.Request) {
		var req agent.CommitRequest
		readJSON(t, r, &req)
		if req.By != "me" {
			t.Errorf("by = %q, want %q", req.By, "me")
		}
		writeJSON(t, w, agent.MediaTypeCommitResponse, agent.CommitResponse{Ref: "img:committed"})
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
	mux.HandleFunc(agent.EndpointDestroy, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, agent.MediaTypeDestroyResponse, agent.DestroyResponse{})
	})
	ctr := serveContainer(t, mux)

	if err := ctr.Destroy(context.Background()); err != nil {
		t.Errorf("destroy: %v", err)
	}
}

func TestContainerExecExitCode(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointExec, func(w http.ResponseWriter, r *http.Request) {
		var req agent.ExecRequest
		readJSON(t, r, &req)
		if req.Command != "false" {
			t.Errorf("command = %q, want %q", req.Command, "false")
		}
		writeExecStream(t, w, agent.ExecResponse{Type: agent.ExecExit, Code: 3})
	})
	ctr := serveContainer(t, mux)

	code, err := ctr.Exec(context.Background(), io.Discard, io.Discard, "false", agent.Options{})
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if code != 3 {
		t.Errorf("code = %d, want 3", code)
	}
}

func TestContainerExecOutput(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointExec, func(w http.ResponseWriter, r *http.Request) {
		writeExecStream(t, w,
			agent.ExecResponse{Type: agent.ExecOutput, Channel: agent.ChannelStdout, Data: []byte("out\xff")},
			agent.ExecResponse{Type: agent.ExecOutput, Channel: agent.ChannelStderr, Data: []byte("err")},
			agent.ExecResponse{Type: agent.ExecExit},
		)
	})
	ctr := serveContainer(t, mux)

	var stdout, stderr bytes.Buffer
	code, err := ctr.Exec(context.Background(), &stdout, &stderr, "x", agent.Options{})
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
	mux.HandleFunc(agent.EndpointExec, func(w http.ResponseWriter, r *http.Request) {
		writeExecStream(t, w, agent.ExecResponse{
			Type:  agent.ExecError,
			Error: &agent.Error{Code: agent.ErrorCodePreconditionFailed, Message: "container not started"},
		})
	})
	ctr := serveContainer(t, mux)

	_, err := ctr.Exec(context.Background(), io.Discard, io.Discard, "x", agent.Options{})
	var opErr *agent.Error
	if !errors.As(err, &opErr) || opErr.Code != agent.ErrorCodePreconditionFailed {
		t.Errorf("exec = %v, want a precondition_failed spec error", err)
	}
}

func TestContainerExecWithoutResult(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointExec, func(w http.ResponseWriter, r *http.Request) {
		writeExecStream(t, w)
	})
	ctr := serveContainer(t, mux)

	_, err := ctr.Exec(context.Background(), io.Discard, io.Discard, "x", agent.Options{})
	if !errors.Is(err, ErrCruxd) {
		t.Errorf("exec = %v, want ErrCruxd", err)
	}
}

func TestContainerExecUnknownChannel(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointExec, func(w http.ResponseWriter, r *http.Request) {
		writeExecStream(t, w, agent.ExecResponse{Type: agent.ExecOutput, Channel: "<bogus>"})
	})
	ctr := serveContainer(t, mux)

	_, err := ctr.Exec(context.Background(), io.Discard, io.Discard, "x", agent.Options{})
	if !errors.Is(err, ErrCruxd) {
		t.Errorf("exec = %v, want ErrCruxd", err)
	}
}
