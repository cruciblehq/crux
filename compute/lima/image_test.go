package lima

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveImageIfStaged(t *testing.T) {
	root := t.TempDir()
	stagingDir := filepath.Join(root, "staging")
	if err := os.MkdirAll(stagingDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(staging): %v", err)
	}
	staged := filepath.Join(stagingDir, "image.qcow2")
	if err := os.WriteFile(staged, []byte("data"), 0o600); err != nil {
		t.Fatalf("WriteFile(staged): %v", err)
	}
	unrelated := filepath.Join(root, "outside.qcow2")
	if err := os.WriteFile(unrelated, []byte("data"), 0o600); err != nil {
		t.Fatalf("WriteFile(unrelated): %v", err)
	}

	if err := RemoveImageIfStaged(context.Background(), staged, stagingDir); err != nil {
		t.Fatalf("RemoveImageIfStaged(staged): %v", err)
	}
	if _, err := os.Stat(staged); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("staged image still exists after cleanup: %v", err)
	}
	if err := RemoveImageIfStaged(context.Background(), unrelated, stagingDir); err != nil {
		t.Fatalf("RemoveImageIfStaged(unrelated): %v", err)
	}
	if _, err := os.Stat(unrelated); err != nil {
		t.Fatalf("unrelated image should remain: %v", err)
	}
}
