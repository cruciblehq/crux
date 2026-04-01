package archive

import (
	"archive/tar"
	"bytes"
	"testing"
)

func TestFind(t *testing.T) {
	// Build a tar archive in memory with two files
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	files := map[string]string{
		"crucible.yaml": "version: 0\n",
		"other.txt":     "hello",
	}
	for name, content := range files {
		tw.WriteHeader(&tar.Header{
			Name:     name,
			Size:     int64(len(content)),
			Mode:     0644,
			Typeflag: tar.TypeReg,
		})
		tw.Write([]byte(content))
	}
	tw.Close()

	t.Run("found", func(t *testing.T) {
		tr := tar.NewReader(bytes.NewReader(buf.Bytes()))
		data, err := Find(tr, "crucible.yaml")
		if err != nil {
			t.Fatalf("FindInTar error: %v", err)
		}
		if string(data) != "version: 0\n" {
			t.Fatalf("got %q, want %q", data, "version: 0\n")
		}
	})

	t.Run("not found", func(t *testing.T) {
		tr := tar.NewReader(bytes.NewReader(buf.Bytes()))
		data, err := Find(tr, "missing.txt")
		if err != nil {
			t.Fatalf("FindInTar error: %v", err)
		}
		if data != nil {
			t.Fatalf("expected nil data, got %q", data)
		}
	})

	t.Run("empty archive", func(t *testing.T) {
		var empty bytes.Buffer
		emptyTw := tar.NewWriter(&empty)
		emptyTw.Close()

		tr := tar.NewReader(bytes.NewReader(empty.Bytes()))
		data, err := Find(tr, "any.txt")
		if err != nil {
			t.Fatalf("FindInTar error: %v", err)
		}
		if data != nil {
			t.Fatalf("expected nil data, got %q", data)
		}
	})
}
