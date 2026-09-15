package local

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestRun_Success(t *testing.T) {
	called := false
	err := run(context.Background(), func(_ context.Context, stdout, stderr io.Writer, command string, args ...string) (int, error) {
		called = true
		if command != "ip" {
			t.Fatalf("command = %q, want %q", command, "ip")
		}
		if got, want := strings.Join(args, " "), "netns exec ns ip addr"; got != want {
			t.Fatalf("args = %q, want %q", got, want)
		}
		return 0, nil
	}, "ip", "netns", "exec", "ns", "ip", "addr")
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !called {
		t.Fatal("run() did not call exec")
	}
}

func TestRun_ExitCodeIsClassified(t *testing.T) {
	err := run(context.Background(), func(_ context.Context, _, _ io.Writer, _ string, _ ...string) (int, error) {
		return 7, nil
	}, "ip", "link", "show")
	if !errors.Is(err, ErrHostExec) {
		t.Fatalf("run() error = %v, want ErrHostExec", err)
	}
}

func TestExecOutput_TrimsWhitespace(t *testing.T) {
	out, err := execOutput(context.Background(), func(_ context.Context, stdout, _ io.Writer, _ string, _ ...string) (int, error) {
		_, _ = stdout.Write([]byte("  hello\n\n"))
		return 0, nil
	}, "echo", "hello")
	if err != nil {
		t.Fatalf("execOutput() error = %v", err)
	}
	if out != "hello" {
		t.Fatalf("execOutput() = %q, want %q", out, "hello")
	}
}
