package cache

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLockedReadCloser_Read(t *testing.T) {
	f := tempFileWithContent(t, "hello world")

	unlocked := false
	rc := &lockedReadCloser{file: f, unlock: func() { unlocked = true }}

	buf := make([]byte, 5)
	n, err := rc.Read(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Fatalf("expected 5 bytes, got %d", n)
	}
	if string(buf) != "hello" {
		t.Fatalf("expected %q, got %q", "hello", string(buf))
	}
	if unlocked {
		t.Fatal("unlock called prematurely")
	}

	rc.Close()
}

func TestLockedReadCloser_ReadAll(t *testing.T) {
	content := "the quick brown fox"
	f := tempFileWithContent(t, content)

	rc := &lockedReadCloser{file: f, unlock: func() {}}

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != content {
		t.Fatalf("expected %q, got %q", content, string(data))
	}

	rc.Close()
}

func TestLockedReadCloser_ReadEmpty(t *testing.T) {
	f := tempFileWithContent(t, "")

	rc := &lockedReadCloser{file: f, unlock: func() {}}

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("expected empty read, got %d bytes", len(data))
	}

	rc.Close()
}

func TestLockedReadCloser_ReadEOF(t *testing.T) {
	f := tempFileWithContent(t, "ab")

	rc := &lockedReadCloser{file: f, unlock: func() {}}

	buf := make([]byte, 10)
	n, err := rc.Read(buf)
	if n != 2 {
		t.Fatalf("expected 2 bytes, got %d", n)
	}

	n, err = rc.Read(buf)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF, got %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 bytes at EOF, got %d", n)
	}

	rc.Close()
}

func TestLockedReadCloser_CloseCallsUnlock(t *testing.T) {
	f := tempFileWithContent(t, "data")

	unlocked := false
	rc := &lockedReadCloser{file: f, unlock: func() { unlocked = true }}

	if err := rc.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !unlocked {
		t.Fatal("unlock was not called on Close")
	}
}

func TestLockedReadCloser_CloseReturnsFileError(t *testing.T) {
	f := tempFileWithContent(t, "data")

	// Close the underlying file first so the lockedReadCloser.Close gets an error.
	f.Close()

	unlocked := false
	rc := &lockedReadCloser{file: f, unlock: func() { unlocked = true }}

	err := rc.Close()
	if err == nil {
		t.Fatal("expected error from closing already-closed file")
	}
	if !unlocked {
		t.Fatal("unlock must be called even when file.Close fails")
	}
}

func TestLockedReadCloser_UnlockCalledOnce(t *testing.T) {
	f := tempFileWithContent(t, "data")

	count := 0
	rc := &lockedReadCloser{file: f, unlock: func() { count++ }}

	rc.Close()
	if count != 1 {
		t.Fatalf("expected unlock to be called once, got %d", count)
	}
}

func TestLockedReadCloser_MultipleReads(t *testing.T) {
	content := "abcdefghij"
	f := tempFileWithContent(t, content)

	rc := &lockedReadCloser{file: f, unlock: func() {}}

	var result bytes.Buffer
	buf := make([]byte, 3)
	for {
		n, err := rc.Read(buf)
		result.Write(buf[:n])
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if result.String() != content {
		t.Fatalf("expected %q, got %q", content, result.String())
	}

	rc.Close()
}

func TestLockedReadCloser_ImplementsReadCloser(t *testing.T) {
	f := tempFileWithContent(t, "data")

	rc := &lockedReadCloser{file: f, unlock: func() {}}

	// Verify the interface is satisfied at compile time via assignment.
	var _ io.ReadCloser = rc

	rc.Close()
}

func tempFileWithContent(t *testing.T, content string) *os.File {
	t.Helper()
	path := filepath.Join(t.TempDir(), "testfile")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open temp file: %v", err)
	}
	return f
}
