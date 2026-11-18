package lima

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cruciblehq/utils-go/file"
)

// Resolves a machine image path, removing the staged copy if it resides in
// the given temp directory.
func RemoveImageIfStaged(ctx context.Context, id, tempDir string) error {
	abs, err := filepath.Abs(id)
	if err != nil {
		return err
	}
	prefix, err := filepath.Abs(tempDir)
	if err != nil {
		return err
	}
	if !file.IsInside(abs, prefix) {
		return nil
	}
	return os.Remove(abs)
}
